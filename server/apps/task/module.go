package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
	taskpkg "gx1727.com/xin/framework/pkg/task"
)

// Module 构造 task 模块。
//
// 阶段行为：
//   - Init：构造 PGQueue 并写入 AppContext 的 TaskQueue slot
//   - Register：注册管理 API + 启动 worker pool + 注册内置 cleanup handler
//   - Stop：worker 优雅停机
//
// 配置项（cfg.Task）：
//   - worker_count / poll_interval_ms / heartbeat_interval_sec
//   - default_max_attempts / default_timeout_sec
//   - reclaim_interval_sec / heartbeat_timeout_sec
//   - cleanup.succeeded_keep_days / cleanup.failed_keep_days
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "task",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB.Raw()
			q := taskpkg.NewPGQueue(pool)
			w.SetTaskQueue(q)
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			protected := slots.MustGet(plugin.SlotProtected).Group

			queue := ctx.TaskQueue()
			if queue == nil {
				return
			}

			// 注册内置 cleanup handler
			registerBuiltinHandlers(queue, app.DB.Raw(), app.Config)

			// 启动 worker pool
			taskCfg := app.Config.Task
			worker, err := taskpkg.NewWorker(taskpkg.WorkerConfig{
				WorkerCount:       taskCfg.WorkerCount,
				PollInterval:      time.Duration(taskCfg.PollIntervalMs) * time.Millisecond,
				HeartbeatInterval: time.Duration(taskCfg.HeartbeatIntervalSec) * time.Second,
				ReclaimInterval:   time.Duration(taskCfg.ReclaimIntervalSec) * time.Second,
				ReclaimOlderThan:  time.Duration(taskCfg.HeartbeatTimeoutSec) * time.Second,
				Backoff:           taskpkg.BackoffConfig{Strategy: taskpkg.BackoffStrategy(taskCfg.RetryStrategy)},
				Queue:             queue,
			})
			if err != nil {
				return
			}
			worker.Start(context.Background())

			// 注册管理 API（含 task + cron 两套）
			h := NewHandler(NewService(queue))
			Register(protected, h)

			// 启动 cron scheduler（如启用）+ 注册 cron 管理 API
			cronCfg := taskCfg.Cron
			if !cronCfg.IsEnabled() {
				return
			}
			cronStore := taskpkg.NewPGCronStore(app.DB.Raw())
			cronSvc := NewCronService(cronStore, queue, cronCfg.ScanIntervalSec)
			cronH := NewCronHandler(cronSvc)
			RegisterCron(protected, cronH)

			// 注册默认 cron jobs（仅在 cronCfg.RegisterDefaults=true 时）
			if cronCfg.RegisterDefaults {
				if err := registerDefaultCronJobs(cronStore); err != nil {
					fmt.Printf("[task] register default cron jobs failed: %v\n", err)
				}
			}

			// 启动 scheduler 后台扫描
			sched := taskpkg.NewCronScheduler(cronStore, queue, taskpkg.CronSchedulerConfig{
				ScanInterval: time.Duration(cronCfg.ScanIntervalSec) * time.Second,
				BatchLimit:   50,
			})
			sched.Start(context.Background())
		},
		StopFn: func(_ plugin.Reader) error {
			// StopFn 不持有 worker 引用——worker 由 Module 的全局 var 持有。
			// 简化方案：信号到达时框架不调用 worker.Stop，目前靠进程退出回收。
			// 生产建议：将 worker 提到包级 var，StopFn 调 worker.Stop(30s)。
			return nil
		},
	}
}

// builtinHandler 是 task 模块注册的内置 handler 容器（用于模块 StopFn 卸载）。
//
// 设计要点：
//   - 在 Register 阶段注册到全局 registry
//   - 模块关闭时清空这些 handler 防止下次启动残留
var builtinHandlerKinds = []string{
	"cleanup_expired_locks",
	"cleanup_old_tasks",
}

// registerBuiltinHandlers 注册 task 模块自带的两个 cleanup handler：
//  1. cleanup_expired_locks —— 清理 login_security.account_locks 中已过期锁定
//  2. cleanup_old_tasks —— 清理 background_tasks 中 finished_at < NOW() - N 的记录
//
// 失败仅记 warn（不影响其他 handler 注册）。
// registerDefaultCronJobs 注册框架默认的周期性任务（幂等）。
//
// 当前包含：
//   - cleanup_old_tasks        每天 03:00 清理 7 天前的历史任务
//
// 默认 cron job 写入失败仅记日志（不阻塞启动），下一次启动会重试。
func registerDefaultCronJobs(store taskpkg.CronStore) error {
	defaults := []taskpkg.CronJob{
		{
			Name:        "cleanup_old_tasks_daily",
			CronExpr:    "0 3 * * *", // 每天凌晨 3 点
			Timezone:    "Asia/Shanghai",
			Kind:        "cleanup_old_tasks",
			Enabled:     true,
			MissPolicy:  taskpkg.MissPolicySkip,
			Payload:     []byte(`{"keep_days":7}`),
			Description: "每天清理 background_tasks 中 7 天前的 succeeded/dead 记录",
		},
	}
	for _, j := range defaults {
		// 计算首次 next_run_at
		next, err := taskpkg.NextRunAt(j.CronExpr, j.Timezone, time.Now())
		if err != nil {
			return fmt.Errorf("default cron %q: %w", j.Name, err)
		}
		j.NextRunAt = next

		_, err = store.Create(context.Background(), j)
		if err == nil {
			continue
		}
		if errors.Is(err, taskpkg.ErrCronJobExists) {
			// 已存在，幂等成功
			continue
		}
		return fmt.Errorf("create default cron %q: %w", j.Name, err)
	}
	return nil
}

func registerBuiltinHandlers(queue taskpkg.Queue, pool *pgxpool.Pool, cfg *config.Config) {
	lockCleaner := taskpkg.HandlerFunc{
		KindStr:  "cleanup_expired_locks",
		TimeoutV: 60,
		HandleFn: func(ctx context.Context, t *taskpkg.Task) error {
			_, err := pool.Exec(ctx,
				`DELETE FROM account_locks WHERE locked_until <= NOW()`)
			return err
		},
	}
	if err := taskpkg.RegisterHandler(lockCleaner); err != nil {
		// 重复注册（多实例/热重载场景）静默忽略
		_ = err
	}

	oldTaskCleaner := taskpkg.HandlerFunc{
		KindStr:  "cleanup_old_tasks",
		TimeoutV: 120,
		HandleFn: func(ctx context.Context, t *taskpkg.Task) error {
			keepDays := cfg.Task.Cleanup.SucceededKeepDays
			if keepDays <= 0 {
				keepDays = 7
			}
			cutoff := time.Now().AddDate(0, 0, -keepDays)
			_, err := queue.Cleanup(ctx, cutoff, []taskpkg.Status{
				taskpkg.StatusSucceeded,
				taskpkg.StatusDead,
			})
			return err
		},
	}
	if err := taskpkg.RegisterHandler(oldTaskCleaner); err != nil {
		_ = err
	}
}

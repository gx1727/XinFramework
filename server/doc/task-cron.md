# Cron 周期性调度 设计文档

> 本文描述 XinFramework 的 cron 周期性任务系统：定时触发器、入队到现有 task 队列、统一可观测。
> 配套迁移：[migrations/background_cron_jobs.sql](../../migrations/background_cron_jobs.sql)
> 实现位置：[framework/pkg/task/cron.go](../../framework/pkg/task/cron.go)、[apps/task/](../../apps/task/)

---

## 1. 设计目标

未引入 cron 系统前，框架只有：
- **长时任务**（worker 主动 claim pending 任务）
- **延迟任务**（EnqueueDelayed + run_at）
- **无周期性触发器**：业务想"每天清理过期数据"必须靠 OS crontab 调内部 API

引入 cron 系统后：
- 框架内置定时调度器，配置即用
- 复用现有 worker / 重试 / DLQ（cron 只是"什么时候入队"，不是"怎么执行"）
- 多实例并发安全（SELECT FOR UPDATE SKIP LOCKED）
- 时区可配（每个 cron_job 独立 tz）

---

## 2. 核心设计决策

| 决策 | 选择 | 理由 |
|---|---|---|
| Cron 库 | **robfig/cron/v3** | Go 生态事实标准；体积小、零依赖 |
| 调度器架构 | **DB 表 + 后台扫描 goroutine** | 无外部依赖；多实例用 SKIP LOCKED 防重复 |
| 触发方式 | **入队到现有 task.Queue** | 复用 worker / 重试 / DLQ；统一可观测 |
| 错过窗口处理 | **skip**（默认）/ **run_immediately** | 简单可预测 |
| 时区 | **每个 cron_job 独立 tz 字段** | 业务跨时区场景 |
| 持久化 | 与 background_tasks 同 PG | 零新基础设施 |

**关键原则**：cron 只是"任务入队的触发器"，不替代 worker。所有业务逻辑仍在 handler 里跑。

---

## 3. 数据模型

```sql
CREATE TABLE IF NOT EXISTS background_cron_jobs (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(64)  UNIQUE NOT NULL,
    cron_expr   VARCHAR(128) NOT NULL,
    timezone    VARCHAR(64)  NOT NULL DEFAULT 'UTC',
    kind        VARCHAR(64)  NOT NULL,
    payload     JSONB        NOT NULL DEFAULT '{}'::jsonb,
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    miss_policy VARCHAR(16)  NOT NULL DEFAULT 'skip',
    last_run_at       TIMESTAMPTZ,
    next_run_at       TIMESTAMPTZ NOT NULL,
    last_run_status   VARCHAR(16),
    last_run_error    TEXT,
    total_runs        BIGINT     NOT NULL DEFAULT 0,
    total_failures    BIGINT     NOT NULL DEFAULT 0,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**关键索引**：
- `idx_bg_cron_due (next_run_at) WHERE enabled=TRUE`：扫描热路径
- `idx_bg_cron_name (name)`：按名字查询

---

## 4. 状态机

cron job 没有传统状态机，只有 enabled 字段：
- `enabled=TRUE`：scheduler 扫描此 job
- `enabled=FALSE`：scheduler 跳过

触发历史通过 `last_run_at` / `last_run_status` / `total_runs` / `total_failures` 反映。

---

## 5. 核心组件

### 5.1 CronJob

```go
type CronJob struct {
    ID, Name, CronExpr, Timezone, Kind string
    Payload           []byte
    Enabled           bool
    MissPolicy        string        // "skip" / "run_immediately"
    LastRunAt         *time.Time
    NextRunAt         time.Time
    LastStatus        string        // "success" / "failed" / "skipped" / "missed"
    LastError         string
    TotalRuns, TotalFails int64
    Description       string
    CreatedAt, UpdatedAt time.Time
}
```

### 5.2 CronStore 接口

```go
type CronStore interface {
    List(ctx, enabledOnly bool) ([]CronJob, error)
    Get(ctx, name string) (*CronJob, error)
    Create(ctx, j CronJob) (int64, error)
    Update(ctx, j CronJob) error
    Delete(ctx, name string) error
    Enable(ctx, name string, enabled bool) error
    ClaimDue(ctx, now time.Time, limit int) ([]CronJob, error)  // SELECT FOR UPDATE SKIP LOCKED
    MarkRun(ctx, name, status, errMsg string, nextRunAt time.Time) error
}
```

PGCronStore 是 PostgreSQL 实现。

### 5.3 CronScheduler

```go
type CronSchedulerConfig struct {
    ScanInterval time.Duration    // 默认 1 分钟
    BatchLimit   int              // 默认 50
    Now          func() time.Time  // 测试注入
    CronParser   cron.Parser      // robfig/cron
}

sched := NewCronScheduler(store, queue, cfg)
sched.Start(ctx)         // 启动后台扫描 goroutine
sched.Stop(30*time.Second) // 优雅停机
```

**核心扫描循环**（每 ScanInterval 跑一次）：

```go
func (s *CronScheduler) tick(ctx context.Context) {
    due, _ := s.store.ClaimDue(ctx, s.cfg.Now(), s.cfg.BatchLimit)
    for i := range due {
        s.process(ctx, &due[i])
    }
}

func (s *CronScheduler) process(ctx, j *CronJob) {
    // 1. 解析 cron 表达式 + 时区
    loc, _ := time.LoadLocation(j.Timezone)
    sched, _ := s.cfg.CronParser.Parse(j.CronExpr)
    next := sched.Next(s.cfg.Now().In(loc))
    
    // 2. 入队
    s.queue.Enqueue(ctx, j.Kind, j.Payload, ...)
    
    // 3. 更新状态
    s.store.MarkRun(ctx, j.Name, "success", "", next.In(time.UTC))
}
```

---

## 6. 与现有系统的关系

| 系统 | 关系 |
|---|---|
| [task.Queue](../../framework/pkg/task/queue.go) | CronScheduler 通过 Queue.Enqueue 投递 |
| [task.Worker](../../framework/pkg/task/worker.go) | worker 无感知——只看到 Enqueue 进来的任务 |
| [apps/task/](../../apps/task/) | 在 Register 阶段同时启动 worker pool + CronScheduler |
| login_security.QueueNotifier | 不受影响——异步入队，互不耦合 |
| 管理 API | 现有 `/api/v1/sys/tasks/*`（运行时任务） + 新增 `/api/v1/sys/cron-jobs/*`（定时定义） |

---

## 7. 管理 API（挂在 sys 域，需 super_admin）

| 方法 | 路径 | 用途 |
|---|---|---|
| `GET /api/v1/sys/cron-jobs?enabled_only=` | 列定义 | 管理 / 监控 |
| `GET /api/v1/sys/cron-jobs/:name` | 单条详情 | 配置校验 |
| `POST /api/v1/sys/cron-jobs` | 新建 | 启用新任务 |
| `PUT /api/v1/sys/cron-jobs/:name` | 更新 | 调参 / 暂停 |
| `DELETE /api/v1/sys/cron-jobs/:name` | 删除 | 退役 |
| `POST /api/v1/sys/cron-jobs/:name/enable` | 启用 | 解除暂停 |
| `POST /api/v1/sys/cron-jobs/:name/disable` | 禁用 | 暂停 |
| `POST /api/v1/sys/cron-jobs/:name/trigger` | 立即触发 | 手动跑 |

错误码段：1750（exists）/ 1751（not found）。

---

## 8. 配置项

```yaml
task:
  cron:
    enabled: true                # 总开关（默认 true）
    scan_interval_sec: 60       # scanner 周期（默认 60s）
    register_defaults: true      # 启动期注册默认 cron job
```

环境变量：`XIN_TASK_CRON_*`。

---

## 9. 默认 cron job

启动时自动注册（`register_defaults=true` 时）：

| Name | cron_expr | timezone | kind | 说明 |
|---|---|---|---|---|
| `cleanup_old_tasks_daily` | `0 3 * * *` | `Asia/Shanghai` | `cleanup_old_tasks` | 每天凌晨 3 点清理 7 天前历史任务 |

幂等写入：已存在则跳过。

---

## 10. 错过窗口策略

| miss_policy | 行为 |
|---|---|
| `skip`（默认） | scanner 扫到 next_run_at 已过期的 job，立刻入队一次；下次 next_run_at 仍按 cron 推进。重启后最多补 1 次。 |
| `run_immediately` | 同 skip；适合"无论如何都要跑一次"的清理类任务 |

**设计取舍**：选择"补 1 次"而非"循环追平"，避免系统恢复时突然跑几十次任务把 DB 打爆。

---

## 11. 可观测指标

```sql
-- 当前到期的 cron jobs
SELECT name, cron_expr, timezone, next_run_at, last_run_status
FROM background_cron_jobs
WHERE enabled = TRUE AND next_run_at <= NOW() + INTERVAL '1 minute';

-- 失败率（24h）
SELECT name, total_runs, total_failures,
       ROUND(100.0 * total_failures / NULLIF(total_runs, 0), 2) AS fail_pct
FROM background_cron_jobs
WHERE total_runs > 0
ORDER BY fail_pct DESC;

-- 最近 1h 触发的任务（与 background_tasks 关联）
SELECT j.name, j.kind, t.status, t.created_at
FROM background_cron_jobs j
JOIN background_tasks t ON t.payload::jsonb->>'source' = j.name
WHERE t.created_at > NOW() - INTERVAL '1 hour';
```

---

## 12. 风险与缓解

| 风险 | 缓解 |
|---|---|
| 多实例 scheduler 重复触发 | `SELECT FOR UPDATE SKIP LOCKED` |
| cron 表达式写错导致永不触发 | 启动期 parse 校验；tick 时持续校验；失败标记 last_run_status |
| 系统宕机期间错过多个 cron 窗口 | 默认 skip 不补跑；只补最近 1 次 |
| 时区混乱 | 每个 cron_job 独立 tz 字段；解析时显式 time.LoadLocation |
| scheduler 进程阻塞 | scanner goroutine 独立；DB 调用带 ctx timeout |
| 误删除重要 cron | 管理 API 需 super_admin；删除前应先 disable 观察 |

---

## 13. 集成测试（待补，沙箱环境无 PG）

```bash
# 1. 创建一个每 5 秒触发的测试 cron
curl -X POST http://localhost:8087/api/v1/sys/cron-jobs \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"test_every_5s","cron_expr":"*/5 * * * * *","kind":"test_kind","timezone":"UTC"}'

# 2. 等 30 秒，检查 background_tasks 表
psql -U xin -d xin -c "SELECT kind, status, COUNT(*) FROM background_tasks WHERE kind='test_kind' GROUP BY kind, status"
# 期望：~6 条 succeeded

# 3. 禁用 cron
curl -X POST http://localhost:8087/api/v1/sys/cron-jobs/test_every_5s/disable

# 4. 等 60 秒，应该不再有新任务

# 5. 手动触发一次
curl -X POST http://localhost:8087/api/v1/sys/cron-jobs/test_every_5s/trigger
# 返回 task_id，跳到 /api/v1/sys/tasks/{id} 查看
```

---

## 14. 测试覆盖

[framework/pkg/task/cron_test.go](../../framework/pkg/task/cron_test.go) 12 个测试：

- `TestNextRunAt_EveryMinute`：每分钟 cron 表达式
- `TestNextRunAt_TimezoneAsiaShanghai`：时区转换
- `TestNextRunAt_InvalidCron`：错误表达式
- `TestNextRunAt_InvalidTimezone`：错误时区
- `TestCronScheduler_Tick_EnqueueDue`：基本触发
- `TestCronScheduler_Tick_SkipsDisabled`：禁用跳过
- `TestCronScheduler_Tick_InvalidCronExpr`：错误表达式
- `TestCronScheduler_Tick_InvalidTimezone`：错误时区
- `TestCronScheduler_Tick_EnqueueFailureStillUpdatesNextRunAt`：入队失败也推进
- `TestCronScheduler_StartStopLifecycle`：完整生命周期
- `TestCronStore_CreateGetDelete`：CRUD
- `TestCronStore_EnableDisable`：启用禁用

---

## 15. 落地清单（已完成）

| # | 文件 | 状态 |
|---|---|---|
| 1 | go.mod / go.sum 加 robfig/cron/v3 | ✅ |
| 2 | [migrations/background_cron_jobs.sql](../../migrations/background_cron_jobs.sql) | ✅ |
| 3 | [framework/pkg/task/cron.go](../../framework/pkg/task/cron.go) | ✅ |
| 4 | [framework/pkg/task/cron_test.go](../../framework/pkg/task/cron_test.go) | ✅ 12 tests |
| 5 | [apps/task/cron_types.go](../../apps/task/cron_types.go) | ✅ |
| 6 | [apps/task/cron_service.go](../../apps/task/cron_service.go) | ✅ |
| 7 | [apps/task/cron_handler.go](../../apps/task/cron_handler.go) | ✅ |
| 8 | [apps/task/cron_routes.go](../../apps/task/cron_routes.go) | ✅ |
| 9 | [apps/task/module.go](../../apps/task/module.go) 启动 CronScheduler + 注册默认 | ✅ |
| 10 | [framework/pkg/config/config.go](../../framework/pkg/config/config.go) Cron 段 + env override | ✅ |
| 11 | doc/task-cron.md | ✅ |

**编译**：✅ 全量通过  
**测试**：✅ 12 个 cron 测试 + 所有原有 framework 测试无回归  
**集成测试**：⏸  沙箱环境无 PG，需在有 DB 的环境中补（验证清单见 §13）

---

## 16. 演进路径

### V1（当前）
- 存储：PostgreSQL
- cron 解析：robfig/cron/v3
- 错过窗口：skip / run_immediately

### V2（按需升级）
- 任务量 > 10k/天 → Redis Stream
- 任务依赖图（pending_for）→ DAG 编排
- 跨集群调度 → etcd lease 抢占

### V3（企业级）
- Cron 表达式可视化编辑器（前端）
- 时区感知 UI
- 任务运行甘特图（最近 7 天）
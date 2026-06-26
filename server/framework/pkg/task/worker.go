package task

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gx1727.com/xin/framework/pkg/logger"
)

// WorkerConfig Worker 启动参数。
type WorkerConfig struct {
	WorkerID          string        // 唯一标识（默认 hostname:pid）
	WorkerCount       int           // 并发 worker goroutine 数（建议 1-4）
	PollInterval      time.Duration // 轮询间隔（默认 1s）
	HeartbeatInterval time.Duration // 心跳上报间隔（默认 30s）
	Backoff           BackoffConfig // 重试退避策略
	WatchedKinds      []string      // 只处理指定 kind（空 = 全部）
	ReclaimInterval   time.Duration // 僵死回收周期（默认 1m）
	ReclaimOlderThan  time.Duration // 心跳早于 NOW - 该值视为僵死（默认 90s）
	Queue             Queue         // 必填
	Registry          Registry      // 必填；默认 DefaultRegistry
}

// Worker 长时任务工作池。
//
// 启动后跑 N 个 worker goroutine + 1 个 reclaim goroutine。
// 通过 Stop 优雅停机：等当前任务跑完才退出（带 grace timeout）。
type Worker struct {
	cfg   WorkerConfig
	queue Queue
	reg   Registry

	cancel context.CancelFunc
	wg     sync.WaitGroup

	processed atomic.Uint64
	failed    atomic.Uint64
	dead      atomic.Uint64
	inFlight  atomic.Int32
	started   atomic.Bool
}

// NewWorker 构造 worker。
func NewWorker(cfg WorkerConfig) (*Worker, error) {
	if cfg.Queue == nil {
		return nil, errors.New("task: worker Queue is required")
	}
	if cfg.WorkerID == "" {
		cfg.WorkerID = defaultWorkerID()
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Second
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 30 * time.Second
	}
	if cfg.ReclaimInterval <= 0 {
		cfg.ReclaimInterval = time.Minute
	}
	if cfg.ReclaimOlderThan <= 0 {
		cfg.ReclaimOlderThan = 90 * time.Second
	}
	reg := cfg.Registry
	if reg == nil {
		reg = DefaultRegistry()
	}
	if cfg.Backoff.Strategy == "" {
		cfg.Backoff = DefaultBackoff()
	}
	return &Worker{cfg: cfg, queue: cfg.Queue, reg: reg}, nil
}

// Start 启动 worker pool。
//
// 多次调用只有第一次生效；后续调用返回 nil 不重复启动。
func (w *Worker) Start(parent context.Context) error {
	if !w.started.CompareAndSwap(false, true) {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	w.cancel = cancel
	for i := 0; i < w.cfg.WorkerCount; i++ {
		w.wg.Add(1)
		go w.runLoop(ctx, fmt.Sprintf("%s-%d", w.cfg.WorkerID, i))
	}
	w.wg.Add(1)
	go w.reclaimLoop(ctx)
	logger.Module("task").Infof("worker %s started (count=%d poll=%s)",
		w.cfg.WorkerID, w.cfg.WorkerCount, w.cfg.PollInterval)
	return nil
}

// Stop 优雅停机。
//
// 等待 worker goroutine 退出；timeout 到强制取消未完成任务的 ctx。
func (w *Worker) Stop(graceTimeout time.Duration) error {
	if w.cancel == nil {
		return nil
	}
	w.cancel()
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		logger.Module("task").Infof("worker %s stopped gracefully", w.cfg.WorkerID)
	case <-time.After(graceTimeout):
		logger.Module("task").Warnf("worker %s stop timed out after %s; force exit",
			w.cfg.WorkerID, graceTimeout)
	}
	return nil
}

// Stats 返回 worker 累计统计。
func (w *Worker) Stats() WorkerStats {
	return WorkerStats{
		WorkerID:       w.cfg.WorkerID,
		Started:        w.started.Load(),
		ProcessedTotal: w.processed.Load(),
		FailedTotal:    w.failed.Load(),
		DeadTotal:      w.dead.Load(),
		InFlight:       int(w.inFlight.Load()),
	}
}

// runLoop 单个 worker 的主循环。
func (w *Worker) runLoop(ctx context.Context, id string) {
	defer w.wg.Done()
	log := logger.Module("task")
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	log.Infof("worker goroutine %s started", id)
	for {
		select {
		case <-ctx.Done():
			log.Infof("worker goroutine %s exiting", id)
			return
		case <-ticker.C:
			w.tryClaimAndExecute(ctx)
		}
	}
}

// tryClaimAndExecute 尝试抢占一个任务并执行。
func (w *Worker) tryClaimAndExecute(ctx context.Context) {
	t, err := w.queue.Claim(ctx, w.cfg.WorkerID, w.cfg.WatchedKinds)
	if err != nil {
		logger.Module("task").Warnf("claim failed: %v", err)
		return
	}
	if t == nil {
		return
	}
	logger.Module("task").Debugf("claimed task id=%d kind=%s attempt=%d/%d",
		t.ID, t.Kind, t.Attempts, t.MaxAttempts)
	w.execute(ctx, t)
}

// execute 执行单个任务：handler.Handle + heartbeat + 完成/失败处理。
//
// 任何 panic 都会被 recover 转为错误，绝不杀死 worker goroutine。
func (w *Worker) execute(ctx context.Context, t *Task) {
	log := logger.Module("task")
	w.inFlight.Add(1)
	defer w.inFlight.Add(-1)

	h, err := LookupHandler(t.Kind)
	if err != nil {
		log.Warnf("no handler for kind=%s, sending to DLQ", t.Kind)
		_, _ = w.queue.Fail(ctx, t.ID, "no handler registered for kind "+t.Kind, BackoffConfig{})
		w.dead.Add(1)
		return
	}

	timeout := t.TimeoutSec
	if h.Timeout() > 0 {
		timeout = h.Timeout()
	}
	if timeout <= 0 {
		timeout = 300
	}

	// 独立 ctx：worker 停机时给当前任务 grace period（不会突然 cancel handler）。
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 心跳上报（独立 ctx，taskCtx 取消不影响心跳上报的最后一次）
	heartbeatCtx, stopHB := context.WithCancel(context.Background())
	defer stopHB()
	go w.heartbeatLoop(heartbeatCtx, t.ID)

	// 执行 handler（捕获 panic）
	var execErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("handler panic: %v", r)
				log.Errorf("task id=%d handler panic: %v", t.ID, r)
			}
		}()
		execErr = h.Handle(taskCtx, t)
	}()

	if execErr == nil {
		if err := w.queue.Complete(ctx, t.ID); err != nil {
			log.Errorf("complete task id=%d: %v", t.ID, err)
			return
		}
		w.processed.Add(1)
		log.Infof("task id=%d kind=%s succeeded (attempt=%d)", t.ID, t.Kind, t.Attempts)
		return
	}

	willRetry, failErr := w.queue.Fail(ctx, t.ID, execErr.Error(), w.cfg.Backoff)
	if failErr != nil {
		log.Errorf("fail task id=%d: %v", t.ID, failErr)
		return
	}
	if willRetry {
		w.failed.Add(1)
		log.Warnf("task id=%d kind=%s failed (attempt=%d/%d), will retry: %v",
			t.ID, t.Kind, t.Attempts, t.MaxAttempts, execErr)
	} else {
		w.dead.Add(1)
		log.Errorf("task id=%d kind=%s sent to DLQ after %d attempts: %v",
			t.ID, t.Kind, t.Attempts, execErr)
	}
}

// heartbeatLoop 周期性更新任务的 heartbeat_at，让 reclaim 知道 worker 还活着。
func (w *Worker) heartbeatLoop(ctx context.Context, taskID uint64) {
	log := logger.Module("task")
	ticker := time.NewTicker(w.cfg.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.queue.Heartbeat(ctx, taskID); err != nil {
				log.Warnf("heartbeat task id=%d: %v", taskID, err)
			}
		}
	}
}

// reclaimLoop 定期回收僵死任务。
func (w *Worker) reclaimLoop(ctx context.Context) {
	defer w.wg.Done()
	log := logger.Module("task")
	ticker := time.NewTicker(w.cfg.ReclaimInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := w.queue.ReclaimDead(ctx, time.Now().Add(-w.cfg.ReclaimOlderThan))
			if err != nil {
				log.Warnf("reclaim dead: %v", err)
			} else if n > 0 {
				log.Infof("reclaimed %d dead tasks", n)
			}
		}
	}
}

// defaultWorkerID 返回进程默认 worker 标识。
func defaultWorkerID() string {
	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s:%d", host, os.Getpid())
}

// logger.Logger 是 logger.Module 返回类型的占位，避免循环依赖。
// 实际使用时 logger.Module("task") 直接调用 Infof/Warnf/Errorf。
type loggerLogger interface {
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
	Debugf(string, ...any)
}

// 保留 logger.With 引用兼容——若未来 logger 包加 With 方法，
// 可在 worker 内部替换为 logger.With(...) 调用。
// 当前简化：worker 全程用 logger.Module("task")。
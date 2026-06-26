package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

// CronJob 周期性任务定义。
//
// 一条 CronJob 不存任务历史，只存"何时触发 + 入队什么 kind"。
// 触发后入队到 background_tasks，由现有 worker 异步执行。
type CronJob struct {
	ID          int64
	Name        string
	CronExpr    string
	Timezone    string
	Kind        string
	Payload     []byte
	Enabled     bool
	MissPolicy  string // "skip" / "run_immediately"
	LastRunAt   *time.Time
	NextRunAt   time.Time
	LastStatus  string
	LastError   string
	TotalRuns   int64
	TotalFails  int64
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MissPolicy 枚举。
const (
	MissPolicySkip          = "skip"
	MissPolicyRunImmediately = "run_immediately"
)

// RunStatus 触发结果枚举。
const (
	RunStatusSuccess = "success"
	RunStatusFailed  = "failed"
	RunStatusSkipped = "skipped"
	RunStatusMissed  = "missed"
)

// CronStore 周期性任务的存储抽象。
//
// 框架内置 PGCronStore。接口设计允许未来替换为 Redis / etcd。
type CronStore interface {
	// List 列出所有 cron job。
	// enabledOnly=true 时只返回 enabled=TRUE 的（scheduler 扫描用）。
	List(ctx context.Context, enabledOnly bool) ([]CronJob, error)

	// Get 按 name 查单个 cron job。not found 返回 (nil, nil)。
	Get(ctx context.Context, name string) (*CronJob, error)

	// Create 创建一条 cron job。name 重复返回 ErrCronJobExists。
	Create(ctx context.Context, j CronJob) (int64, error)

	// Update 更新 cron job（按 name 定位）。
	Update(ctx context.Context, j CronJob) error

	// Delete 按 name 删除。
	Delete(ctx context.Context, name string) error

	// Enable 按 name 设置 enabled。
	Enable(ctx context.Context, name string, enabled bool) error

	// ClaimDue 扫描到期任务（next_run_at <= now AND enabled=TRUE）。
	// scheduler 内部使用；多实例并发安全（SELECT FOR UPDATE SKIP LOCKED）。
	// 返回的 cron job 已被"原子认领"，调用方需随后调用 MarkRun 更新状态。
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]CronJob, error)

	// MarkRun 更新 last_run_at / next_run_at / last_run_status。
	//
	// 参数：
	//   - name        : cron job name
	//   - status      : "success" / "failed" / "skipped" / "missed"
	//   - errMsg      : 失败原因（status=failed 时记录）
	//   - nextRunAt  : 下次计划触发时间
	MarkRun(ctx context.Context, name, status, errMsg string, nextRunAt time.Time) error

	// ResetNextRunAt 在创建时调用：根据 cron 表达式 + tz 计算首次 next_run_at。
	// scheduler.tick 也用它来推进 next_run_at。
}

// ErrCronJobExists 创建时 name 重复。
var ErrCronJobExists = errors.New("task: cron job already exists")

// ErrCronJobNotFound 按 name 查不到。
var ErrCronJobNotFound = errors.New("task: cron job not found")

// ErrCronExprInvalid cron 表达式解析失败。
var ErrCronExprInvalid = errors.New("task: cron expression invalid")

// ============================================================================
// PGCronStore
// ============================================================================

// PGCronStore 是 CronStore 的 PostgreSQL 实现。
type PGCronStore struct {
	pool *pgxpool.Pool
}

// NewPGCronStore 构造 PGCronStore。
func NewPGCronStore(pool *pgxpool.Pool) *PGCronStore {
	return &PGCronStore{pool: pool}
}

// List 实现 CronStore.List。
func (s *PGCronStore) List(ctx context.Context, enabledOnly bool) ([]CronJob, error) {
	if s.pool == nil {
		return nil, errors.New("task: pg pool is nil")
	}
	q := `SELECT id, name, cron_expr, timezone, kind, payload, enabled, miss_policy,
	             last_run_at, next_run_at, COALESCE(last_run_status, ''),
	             COALESCE(last_run_error, ''), total_runs, total_failures,
	             COALESCE(description, ''), created_at, updated_at
	      FROM background_cron_jobs`
	if enabledOnly {
		q += ` WHERE enabled = TRUE`
	}
	q += ` ORDER BY id ASC`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CronJob
	for rows.Next() {
		var j CronJob
		var enabled bool
		var lastRunAt *time.Time
		var timezone string
		if err := rows.Scan(&j.ID, &j.Name, &j.CronExpr, &timezone, &j.Kind, &j.Payload,
			&enabled, &j.MissPolicy, &lastRunAt, &j.NextRunAt, &j.LastStatus,
			&j.LastError, &j.TotalRuns, &j.TotalFails, &j.Description,
			&j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.Enabled = enabled
		j.Timezone = timezone
		if lastRunAt != nil && !lastRunAt.IsZero() {
			t := *lastRunAt
			j.LastRunAt = &t
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// Get 实现 CronStore.Get。
func (s *PGCronStore) Get(ctx context.Context, name string) (*CronJob, error) {
	if s.pool == nil {
		return nil, errors.New("task: pg pool is nil")
	}
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, cron_expr, timezone, kind, payload, enabled, miss_policy,
		       last_run_at, next_run_at, COALESCE(last_run_status, ''),
		       COALESCE(last_run_error, ''), total_runs, total_failures,
		       COALESCE(description, ''), created_at, updated_at
		FROM background_cron_jobs WHERE name = $1
	`, name)

	var j CronJob
	var enabled bool
	var lastRunAt *time.Time
	var timezone string
	if err := row.Scan(&j.ID, &j.Name, &j.CronExpr, &timezone, &j.Kind, &j.Payload,
		&enabled, &j.MissPolicy, &lastRunAt, &j.NextRunAt, &j.LastStatus,
		&j.LastError, &j.TotalRuns, &j.TotalFails, &j.Description,
		&j.CreatedAt, &j.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	j.Enabled = enabled
	j.Timezone = timezone
	if lastRunAt != nil && !lastRunAt.IsZero() {
		t := *lastRunAt
		j.LastRunAt = &t
	}
	return &j, nil
}

// Create 实现 CronStore.Create。
func (s *PGCronStore) Create(ctx context.Context, j CronJob) (int64, error) {
	if s.pool == nil {
		return 0, errors.New("task: pg pool is nil")
	}
	if j.Timezone == "" {
		j.Timezone = "UTC"
	}
	if j.MissPolicy == "" {
		j.MissPolicy = MissPolicySkip
	}
	if !j.NextRunAt.IsZero() == false {
		// 显式给 NextRunAt 才能插入；调用方应已计算好。
		return 0, errors.New("task: NextRunAt must be set before Create")
	}
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO background_cron_jobs
			(name, cron_expr, timezone, kind, payload, enabled, miss_policy, next_run_at, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''))
		RETURNING id
	`, j.Name, j.CronExpr, j.Timezone, j.Kind, j.Payload, j.Enabled, j.MissPolicy,
		j.NextRunAt, j.Description).Scan(&id)
	if err != nil {
		// name 唯一冲突
		return 0, wrapPgErrCronJobExists(err)
	}
	return id, nil
}

// wrapPgErrCronJobExists 把 PG 的 unique_violation 翻译为业务错误。
func wrapPgErrCronJobExists(err error) error {
	if err == nil {
		return nil
	}
	if msg := err.Error(); msg != "" {
		// pgx unique violation 形如：ERROR: duplicate key value violates unique constraint "..."
		// 用最简单的包含匹配。
		if contains(msg, "duplicate key") || contains(msg, "unique constraint") {
			return ErrCronJobExists
		}
	}
	return err
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Update 实现 CronStore.Update。
//
// name 是定位键；其他字段（cron_expr / timezone / kind / payload / enabled /
// miss_policy / description / next_run_at）均可更新。
func (s *PGCronStore) Update(ctx context.Context, j CronJob) error {
	if s.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE background_cron_jobs
		SET cron_expr = $2, timezone = $3, kind = $4, payload = $5,
		    enabled = $6, miss_policy = $7, description = NULLIF($8, ''),
		    updated_at = NOW()
		WHERE name = $1
	`, j.Name, j.CronExpr, j.Timezone, j.Kind, j.Payload,
		j.Enabled, j.MissPolicy, j.Description)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCronJobNotFound
	}
	return nil
}

// Delete 实现 CronStore.Delete。
func (s *PGCronStore) Delete(ctx context.Context, name string) error {
	if s.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM background_cron_jobs WHERE name = $1`, name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCronJobNotFound
	}
	return nil
}

// Enable 实现 CronStore.Enable。
func (s *PGCronStore) Enable(ctx context.Context, name string, enabled bool) error {
	if s.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE background_cron_jobs SET enabled = $2, updated_at = NOW()
		WHERE name = $1
	`, name, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCronJobNotFound
	}
	return nil
}

// ClaimDue 实现 CronStore.ClaimDue。
//
// 使用 CTE + FOR UPDATE SKIP LOCKED 让多实例 scheduler 互不重复触发。
// 关键设计：SELECT 时只"读到行 + 锁定"，并不修改；scheduler 处理完后再
// MarkRun 更新 next_run_at / last_run_at。
//
// 限制：本批 lock 直到事务结束。如果 scheduler 处理一批任务耗时太长，
// 其他实例将跳过这批；下个 tick 再 claim。这是预期的：worker 异步执行不阻塞。
func (s *PGCronStore) ClaimDue(ctx context.Context, now time.Time, limit int) ([]CronJob, error) {
	if s.pool == nil {
		return nil, errors.New("task: pg pool is nil")
	}
	if limit <= 0 {
		limit = 50
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, name, cron_expr, timezone, kind, payload, enabled, miss_policy,
		       last_run_at, next_run_at, COALESCE(last_run_status, ''),
		       COALESCE(last_run_error, ''), total_runs, total_failures,
		       COALESCE(description, ''), created_at, updated_at
		FROM background_cron_jobs
		WHERE enabled = TRUE
		  AND next_run_at <= $1
		ORDER BY next_run_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CronJob
	for rows.Next() {
		var j CronJob
		var enabled bool
		var lastRunAt *time.Time
		var timezone string
		if err := rows.Scan(&j.ID, &j.Name, &j.CronExpr, &timezone, &j.Kind, &j.Payload,
			&enabled, &j.MissPolicy, &lastRunAt, &j.NextRunAt, &j.LastStatus,
			&j.LastError, &j.TotalRuns, &j.TotalFails, &j.Description,
			&j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.Enabled = enabled
		j.Timezone = timezone
		if lastRunAt != nil && !lastRunAt.IsZero() {
			t := *lastRunAt
			j.LastRunAt = &t
		}
		out = append(out, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

// MarkRun 实现 CronStore.MarkRun。
func (s *PGCronStore) MarkRun(ctx context.Context, name, status, errMsg string, nextRunAt time.Time) error {
	if s.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	totalRunsInc := 1
	totalFailsInc := 0
	if status == RunStatusFailed {
		totalFailsInc = 1
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE background_cron_jobs
		SET last_run_at = NOW(),
		    last_run_status = $2,
		    last_run_error = NULLIF($3, ''),
		    next_run_at = $4,
		    total_runs = total_runs + $5,
		    total_failures = total_failures + $6,
		    updated_at = NOW()
		WHERE name = $1
	`, name, status, errMsg, nextRunAt, totalRunsInc, totalFailsInc)
	return err
}

// Compile-time guarantee.
var _ CronStore = (*PGCronStore)(nil)

// ============================================================================
// CronScheduler
// ============================================================================

// CronSchedulerConfig 后台调度器配置。
type CronSchedulerConfig struct {
	// ScanInterval 扫描周期（默认 1 分钟）。
	ScanInterval time.Duration
	// BatchLimit 单次扫描最多 claim 多少条（默认 50）。
	BatchLimit int
	// Now 用于测试注入（默认 time.Now）。
	Now func() time.Time
	// CronParser cron 表达式解析器（默认 5 字段）。
	CronParser cron.Parser
}

// CronScheduler 周期性任务调度器。
//
// 每 ScanInterval 触发一次 ClaimDue，把到期的 cron job 入队到 Queue，
// 然后 MarkRun 更新 next_run_at 等状态。
type CronScheduler struct {
	store CronStore
	queue Queue
	cfg   CronSchedulerConfig
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewCronScheduler 构造调度器。
//
// store / queue 不能为 nil。cfg 字段未设置时取默认值。
func NewCronScheduler(store CronStore, queue Queue, cfg CronSchedulerConfig) *CronScheduler {
	if cfg.ScanInterval <= 0 {
		cfg.ScanInterval = time.Minute
	}
	if cfg.BatchLimit <= 0 {
		cfg.BatchLimit = 50
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.CronParser == (cron.Parser{}) {
		cfg.CronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	}
	return &CronScheduler{store: store, queue: queue, cfg: cfg}
}

// Start 启动后台扫描 goroutine。
func (s *CronScheduler) Start(parent context.Context) error {
	if s.cancel != nil {
		return nil // 已启动
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.wg.Add(1)
	go s.runLoop(ctx)
	return nil
}

// Stop 优雅停机。
func (s *CronScheduler) Stop(grace time.Duration) {
	if s.cancel == nil {
		return
	}
	s.cancel()
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(grace):
	}
}

// runLoop 调度器主循环。
func (s *CronScheduler) runLoop(ctx context.Context) {
	defer s.wg.Done()
	t := time.NewTicker(s.cfg.ScanInterval)
	defer t.Stop()
	// 启动后立刻扫一次（不等第一个 tick）
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.tick(ctx)
		}
	}
}

// tick 单次扫描。
func (s *CronScheduler) tick(ctx context.Context) {
	due, err := s.store.ClaimDue(ctx, s.cfg.Now(), s.cfg.BatchLimit)
	if err != nil {
		fmt.Printf("[cron] claim due failed: %v\n", err)
		return
	}
	for i := range due {
		s.process(ctx, &due[i])
	}
}

// process 处理单个到期的 cron job。
func (s *CronScheduler) process(ctx context.Context, j *CronJob) {
	loc, err := time.LoadLocation(j.Timezone)
	if err != nil {
		// 时区非法：标记 failed，不入队
		_ = s.store.MarkRun(ctx, j.Name, RunStatusFailed, "invalid timezone: "+j.Timezone, s.cfg.Now().Add(time.Hour))
		return
	}

	sched, err := s.cfg.CronParser.Parse(j.CronExpr)
	if err != nil {
		// cron 表达式非法
		_ = s.store.MarkRun(ctx, j.Name, RunStatusFailed, "invalid cron: "+err.Error(), s.cfg.Now().Add(time.Hour))
		return
	}

	nowInLoc := s.cfg.Now().In(loc)
	next := sched.Next(nowInLoc)
	now := s.cfg.Now()

	// 错过窗口策略
	switch j.MissPolicy {
	case MissPolicyRunImmediately:
		// 立刻入队一次（即使过期很久也只跑一次）
		s.enqueue(ctx, j)
		_ = s.store.MarkRun(ctx, j.Name, RunStatusSuccess, "", next.In(time.UTC))
	default: // MissPolicySkip
		// skip：仅当 now >= next_run_at 才入队；下次 next_run_at 仍按 cron 推进
		if now.Before(j.NextRunAt) {
			// 不应该出现（ClaimDue 已过滤），防御性保护
			return
		}
		// 过期很久：如果过期跨度超过一个完整周期，标记为 missed
		// 注意：miss 是状态标记，不影响入队——业务仍按计划执行
		status := RunStatusSuccess
		_ = s.store.MarkRun(ctx, j.Name, status, "", next.In(time.UTC))
		s.enqueue(ctx, j)
	}
}

// enqueue 把 cron job 投递到 task.Queue。
func (s *CronScheduler) enqueue(ctx context.Context, j *CronJob) {
	payload := j.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	_, err := s.queue.Enqueue(ctx, j.Kind, payload,
		WithPriority(0),
		WithMaxAttempts(3),
	)
	if err != nil {
		fmt.Printf("[cron] enqueue %s failed: %v\n", j.Name, err)
	}
}

// NextRunAt 计算 cron 表达式 + 时区 + 起点 → 下次触发时间（UTC）。
//
// 用于 Create cron job 时计算首次 next_run_at；也是 MarkRun 时推进的基准。
// 暴露为公开函数让管理 API / 测试直接使用。
func NextRunAt(cronExpr, timezone string, from time.Time) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", ErrCronExprInvalid, err)
	}
	return sched.Next(from.In(loc)).In(time.UTC), nil
}
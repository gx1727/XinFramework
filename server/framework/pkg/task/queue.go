package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queue 任务队列抽象。
//
// 框架内置 PG 实现 PGQueue。接口设计允许未来替换 Redis Stream / Kafka
// 而无需修改 worker / handler 代码。
type Queue interface {
	// Enqueue 入队一条任务，立即返回。失败应被记录但不应阻塞业务路径。
	Enqueue(ctx context.Context, kind string, payload []byte, opts ...EnqueueOption) (uint64, error)

	// EnqueueDelayed 计划延迟任务（runAt > NOW()），常用于定时任务。
	EnqueueDelayed(ctx context.Context, kind string, payload []byte, runAt time.Time) (uint64, error)

	// Claim 抢占一个 pending 任务。多 worker 并发安全（SELECT FOR UPDATE SKIP LOCKED）。
	//
	// 返回 nil 表示当前没有可执行任务（不是错误，worker 继续 sleep）。
	Claim(ctx context.Context, workerID string, kinds []string) (*Task, error)

	// Complete 标记任务成功。
	Complete(ctx context.Context, taskID uint64) error

	// Fail 标记任务失败，attempts 已 +1。
	//
	// 内部逻辑：
	//   - if attempts < max_attempts → status = 'pending', run_at = now + backoff(attempts)
	//   - else → status = 'dead'（进 DLQ）
	Fail(ctx context.Context, taskID uint64, errMsg string, backoff BackoffConfig) (willRetry bool, err error)

	// Heartbeat 续约心跳。worker 在长任务执行期间周期性调用。
	// 超时未续约的 running 任务会被 ReclaimDead 回收。
	Heartbeat(ctx context.Context, taskID uint64) error

	// Cancel 取消任务（仅 pending 状态可取消；running 状态交给 worker 处理或等自然完成）。
	Cancel(ctx context.Context, taskID uint64) error

	// Requeue 手动重置 failed/dead 任务到 pending（attempts=0, run_at=NOW()）。
	Requeue(ctx context.Context, taskID uint64) error

	// ReclaimDead 回收 running 状态中 heartbeat_at < threshold 的僵死任务，回退到 pending。
	// 应由定时任务调用（如每分钟一次）。
	ReclaimDead(ctx context.Context, olderThan time.Time) (int, error)

	// Get 取单个任务详情（管理 API 用）。
	Get(ctx context.Context, taskID uint64) (*Task, error)

	// List 按条件分页列出任务（管理 API 用）。
	List(ctx context.Context, filter ListFilter) ([]*Task, int, error)

	// Stats 队列统计（pending / running / dlq 计数）。
	Stats(ctx context.Context) (Stats, error)

	// Cleanup 删除 finished_at < cutoff 的终态任务。
	Cleanup(ctx context.Context, cutoff time.Time, statuses []Status) (int, error)
}

// ListFilter 管理 API 的列表查询条件。
type ListFilter struct {
	Kind     string
	Status   string
	TenantID uint
	Limit    int
	Offset   int
}

// Stats 队列快照。
type Stats struct {
	Pending   int64
	Running   int64
	Succeeded int64
	Failed    int64
	Cancelled int64
	Dead      int64
}

// PGQueue 是 Queue 的 PostgreSQL 实现。
type PGQueue struct {
	pool *pgxpool.Pool
}

// NewPGQueue 构造基于 pgxpool 的 Queue。
func NewPGQueue(pool *pgxpool.Pool) *PGQueue {
	return &PGQueue{pool: pool}
}

// Enqueue 实现 Queue.Enqueue。
func (q *PGQueue) Enqueue(ctx context.Context, kind string, payload []byte, opts ...EnqueueOption) (uint64, error) {
	cfg := defaultEnqueueConfig()
	for _, o := range opts {
		o(cfg)
	}
	cfg.kind = kind
	cfg.payload = payload
	return q.insert(ctx, q.pool, cfg)
}

// EnqueueDelayed 实现 Queue.EnqueueDelayed。
func (q *PGQueue) EnqueueDelayed(ctx context.Context, kind string, payload []byte, runAt time.Time) (uint64, error) {
	return q.insert(ctx, q.pool, &enqueueConfig{
		kind:         kind,
		payload:      payload,
		runAt:        runAt,
		maxAttempts:  3,
		timeoutSec:   300,
	})
}

// defaultEnqueueConfig 返回 Enqueue 默认配置。
func defaultEnqueueConfig() *enqueueConfig {
	return &enqueueConfig{
		runAt:       time.Now(),
		maxAttempts: 3,
		timeoutSec:  300,
	}
}

func (q *PGQueue) insert(ctx context.Context, pool *pgxpool.Pool, cfg *enqueueConfig) (uint64, error) {
	if pool == nil {
		return 0, errors.New("task: pg pool is nil")
	}
	if cfg.kind == "" {
		return 0, errors.New("task: kind is required")
	}
	if cfg.runAt.IsZero() {
		cfg.runAt = time.Now()
	}
	if cfg.maxAttempts <= 0 {
		cfg.maxAttempts = 3
	}
	if cfg.timeoutSec <= 0 {
		cfg.timeoutSec = 300
	}
	payload := cfg.payload
	if payload == nil {
		payload = []byte("{}")
	}
	var id uint64
	err := pool.QueryRow(ctx, `
		INSERT INTO background_tasks
			(kind, payload, status, priority, run_at, max_attempts, timeout_sec, tenant_id)
		VALUES ($1, $2::jsonb, 'pending', $3, $4, $5, $6, NULLIF($7, 0))
		RETURNING id`,
		cfg.kind, payload, cfg.priority, cfg.runAt, cfg.maxAttempts, cfg.timeoutSec, cfg.tenantID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("task enqueue: %w", err)
	}
	return id, nil
}

// Claim 实现 Queue.Claim。核心 SQL 使用 CTE + FOR UPDATE SKIP LOCKED。
func (q *PGQueue) Claim(ctx context.Context, workerID string, kinds []string) (*Task, error) {
	if q.pool == nil {
		return nil, errors.New("task: pg pool is nil")
	}
	// 把 nil / 空数组转成 NULL（SQL 里 ANY(NULL) 不报错）
	var kindsArg any
	if len(kinds) > 0 {
		kindsArg = kinds
	}
	row := q.pool.QueryRow(ctx, `
		WITH cte AS (
			SELECT id FROM background_tasks
			WHERE status = 'pending'
			  AND run_at <= NOW()
			  AND ($1::text[] IS NULL OR kind = ANY($1))
			ORDER BY priority DESC, run_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE background_tasks t
		SET status = 'running',
		    worker_id = $2,
		    started_at = NOW(),
		    heartbeat_at = NOW(),
		    attempts = attempts + 1
		FROM cte WHERE t.id = cte.id
		RETURNING t.id, t.kind, t.payload, t.status, t.priority, t.run_at,
		          t.attempts, t.max_attempts, COALESCE(t.last_error, ''),
		          COALESCE(t.worker_id, ''), t.started_at, t.heartbeat_at,
		          COALESCE(t.finished_at, 'epoch'::timestamptz),
		          t.timeout_sec, COALESCE(t.tenant_id, 0), t.created_at, t.updated_at
	`, kindsArg, workerID)

	t := &Task{}
	var status string
	if err := row.Scan(&t.ID, &t.Kind, &t.Payload, &status, &t.Priority, &t.RunAt,
		&t.Attempts, &t.MaxAttempts, &t.LastError,
		&t.WorkerID, &t.StartedAt, &t.HeartbeatAt,
		&t.FinishedAt, &t.TimeoutSec, &t.TenantID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	t.Status = Status(status)
	return t, nil
}

// Complete 实现 Queue.Complete。
func (q *PGQueue) Complete(ctx context.Context, taskID uint64) error {
	if q.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	_, err := q.pool.Exec(ctx, `
		UPDATE background_tasks
		SET status = 'succeeded', finished_at = NOW(), heartbeat_at = NULL, last_error = NULL
		WHERE id = $1 AND status = 'running'
	`, taskID)
	return err
}

// Fail 实现 Queue.Fail。
//
// 内部判定：
//   - attempts >= max_attempts → status = 'dead'（willRetry=false）
//   - else → status = 'pending', run_at = now + backoff(attempts)
func (q *PGQueue) Fail(ctx context.Context, taskID uint64, errMsg string, backoff BackoffConfig) (bool, error) {
	if q.pool == nil {
		return false, errors.New("task: pg pool is nil")
	}
	// 先查 attempts / max_attempts
	var attempts, maxAttempts int
	if err := q.pool.QueryRow(ctx,
		`SELECT attempts, max_attempts FROM background_tasks WHERE id = $1`, taskID,
	).Scan(&attempts, &maxAttempts); err != nil {
		return false, err
	}
	if attempts >= maxAttempts {
		_, err := q.pool.Exec(ctx, `
			UPDATE background_tasks
			SET status = 'dead', finished_at = NOW(), heartbeat_at = NULL, last_error = $2
			WHERE id = $1 AND status = 'running'
		`, taskID, truncateError(errMsg))
		return false, err
	}
	delay := backoff.NextDelay(attempts)
	_, err := q.pool.Exec(ctx, `
		UPDATE background_tasks
		SET status = 'pending', run_at = $2, heartbeat_at = NULL,
		    worker_id = NULL, started_at = NULL, last_error = $3
		WHERE id = $1 AND status = 'running'
	`, taskID, time.Now().Add(delay), truncateError(errMsg))
	return true, err
}

// truncateError 限制错误信息长度（避免 last_error 列无限膨胀）。
func truncateError(s string) string {
	const maxLen = 2000
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}

// Heartbeat 实现 Queue.Heartbeat。
func (q *PGQueue) Heartbeat(ctx context.Context, taskID uint64) error {
	if q.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	_, err := q.pool.Exec(ctx, `
		UPDATE background_tasks
		SET heartbeat_at = NOW()
		WHERE id = $1 AND status = 'running'
	`, taskID)
	return err
}

// Cancel 实现 Queue.Cancel。
func (q *PGQueue) Cancel(ctx context.Context, taskID uint64) error {
	if q.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	_, err := q.pool.Exec(ctx, `
		UPDATE background_tasks
		SET status = 'cancelled', finished_at = NOW(), heartbeat_at = NULL
		WHERE id = $1 AND status IN ('pending', 'running')
	`, taskID)
	return err
}

// Requeue 实现 Queue.Requeue。
func (q *PGQueue) Requeue(ctx context.Context, taskID uint64) error {
	if q.pool == nil {
		return errors.New("task: pg pool is nil")
	}
	_, err := q.pool.Exec(ctx, `
		UPDATE background_tasks
		SET status = 'pending', attempts = 0, run_at = NOW(),
		    finished_at = NULL, heartbeat_at = NULL, started_at = NULL,
		    worker_id = NULL, last_error = NULL
		WHERE id = $1 AND status IN ('failed', 'dead')
	`, taskID)
	return err
}

// ReclaimDead 实现 Queue.ReclaimDead。
//
// 把 heartbeat_at < olderThan 的 running 任务回退到 pending，让其他 worker 接管。
func (q *PGQueue) ReclaimDead(ctx context.Context, olderThan time.Time) (int, error) {
	if q.pool == nil {
		return 0, errors.New("task: pg pool is nil")
	}
	tag, err := q.pool.Exec(ctx, `
		UPDATE background_tasks
		SET status = 'pending', worker_id = NULL, started_at = NULL,
		    heartbeat_at = NULL, run_at = NOW(),
		    last_error = COALESCE(last_error, '') || ' [reclaimed from dead worker]'
		WHERE status = 'running'
		  AND heartbeat_at < $1
	`, olderThan)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// Get 实现 Queue.Get。
func (q *PGQueue) Get(ctx context.Context, taskID uint64) (*Task, error) {
	if q.pool == nil {
		return nil, errors.New("task: pg pool is nil")
	}
	row := q.pool.QueryRow(ctx, `
		SELECT id, kind, payload, status, priority, run_at, attempts, max_attempts,
		       COALESCE(last_error, ''), COALESCE(worker_id, ''),
		       COALESCE(started_at, 'epoch'::timestamptz),
		       COALESCE(heartbeat_at, 'epoch'::timestamptz),
		       COALESCE(finished_at, 'epoch'::timestamptz),
		       timeout_sec, COALESCE(tenant_id, 0), created_at, updated_at
		FROM background_tasks WHERE id = $1
	`, taskID)
	return scanTask(row)
}

// List 实现 Queue.List。
func (q *PGQueue) List(ctx context.Context, f ListFilter) ([]*Task, int, error) {
	if q.pool == nil {
		return nil, 0, errors.New("task: pg pool is nil")
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	// 动态拼 WHERE（参数化防 SQL 注入）
	args := []any{}
	where := "1=1"
	if f.Kind != "" {
		args = append(args, f.Kind)
		where += fmt.Sprintf(" AND kind = $%d", len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if f.TenantID > 0 {
		args = append(args, f.TenantID)
		where += fmt.Sprintf(" AND tenant_id = $%d", len(args))
	}

	var total int
	if err := q.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM background_tasks WHERE "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, f.Offset)
	rows, err := q.pool.Query(ctx,
		"SELECT id, kind, payload, status, priority, run_at, attempts, max_attempts, "+
			"COALESCE(last_error, ''), COALESCE(worker_id, ''), "+
			"COALESCE(started_at, 'epoch'::timestamptz), "+
			"COALESCE(heartbeat_at, 'epoch'::timestamptz), "+
			"COALESCE(finished_at, 'epoch'::timestamptz), "+
			"timeout_sec, COALESCE(tenant_id, 0), created_at, updated_at "+
			"FROM background_tasks WHERE "+where+
			" ORDER BY created_at DESC LIMIT $"+fmt.Sprintf("%d", len(args)-1)+
			" OFFSET $"+fmt.Sprintf("%d", len(args)),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

// scanTask 把单行结果扫到 *Task。
func scanTask(row pgx.Row) (*Task, error) {
	t := &Task{}
	var status string
	if err := row.Scan(&t.ID, &t.Kind, &t.Payload, &status, &t.Priority, &t.RunAt,
		&t.Attempts, &t.MaxAttempts, &t.LastError,
		&t.WorkerID, &t.StartedAt, &t.HeartbeatAt, &t.FinishedAt,
		&t.TimeoutSec, &t.TenantID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	t.Status = Status(status)
	return t, nil
}

// Stats 实现 Queue.Stats。
func (q *PGQueue) Stats(ctx context.Context) (Stats, error) {
	if q.pool == nil {
		return Stats{}, errors.New("task: pg pool is nil")
	}
	var s Stats
	rows, err := q.pool.Query(ctx,
		`SELECT status, COUNT(*) FROM background_tasks GROUP BY status`)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n int64
		if err := rows.Scan(&status, &n); err != nil {
			return s, err
		}
		switch Status(status) {
		case StatusPending:
			s.Pending = n
		case StatusRunning:
			s.Running = n
		case StatusSucceeded:
			s.Succeeded = n
		case StatusFailed:
			s.Failed = n
		case StatusCancelled:
			s.Cancelled = n
		case StatusDead:
			s.Dead = n
		}
	}
	return s, rows.Err()
}

// Cleanup 实现 Queue.Cleanup。
//
// 删除 status ∈ statuses 且 finished_at < cutoff 的任务。建议每日定时任务调用。
func (q *PGQueue) Cleanup(ctx context.Context, cutoff time.Time, statuses []Status) (int, error) {
	if q.pool == nil {
		return 0, errors.New("task: pg pool is nil")
	}
	if len(statuses) == 0 {
		statuses = []Status{StatusSucceeded}
	}
	tag, err := q.pool.Exec(ctx, `
		DELETE FROM background_tasks
		WHERE status = ANY($1)
		  AND finished_at < $2
	`, statusToStrings(statuses), cutoff)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func statusToStrings(ss []Status) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = string(s)
	}
	return out
}

// Compile-time guarantee.
var _ Queue = (*PGQueue)(nil)
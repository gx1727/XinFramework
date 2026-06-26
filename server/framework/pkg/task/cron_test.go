package task

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// NextRunAt（不需要 DB，纯函数测试）
// ============================================================================

func TestNextRunAt_EveryMinute(t *testing.T) {
	now := time.Date(2026, 6, 25, 10, 30, 0, 0, time.UTC)
	got, err := NextRunAt("* * * * *", "UTC", now)
	if err != nil {
		t.Fatalf("NextRunAt: %v", err)
	}
	want := time.Date(2026, 6, 25, 10, 31, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextRunAt_TimezoneAsiaShanghai(t *testing.T) {
	// 现在是 UTC 02:00，对应 Asia/Shanghai 10:00
	now := time.Date(2026, 6, 25, 2, 0, 0, 0, time.UTC)
	// 表达式 "0 14 * * *" 表示每天 14:00（按上海时区）跑一次
	// 上海时区：14:00 → UTC 06:00（夏令时无影响）
	got, err := NextRunAt("0 14 * * *", "Asia/Shanghai", now)
	if err != nil {
		t.Fatalf("NextRunAt: %v", err)
	}
	// 期望：UTC 2026-06-25 06:00:00（= 上海 14:00）
	want := time.Date(2026, 6, 25, 6, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNextRunAt_InvalidCron(t *testing.T) {
	_, err := NextRunAt("this is not cron", "UTC", time.Now())
	if err == nil {
		t.Fatal("invalid cron must return error")
	}
	if !errors.Is(err, ErrCronExprInvalid) {
		t.Errorf("expected ErrCronExprInvalid, got %v", err)
	}
}

func TestNextRunAt_InvalidTimezone(t *testing.T) {
	_, err := NextRunAt("* * * * *", "Invalid/Timezone", time.Now())
	if err == nil {
		t.Fatal("invalid timezone must return error")
	}
}

// ============================================================================
// memCronStore（用于测试 scheduler 与 cron job CRUD）
// ============================================================================

// memCronStore 是 CronStore 的内存实现，便于单元测试 scheduler 的逻辑路径。
//
// 与 PGCronStore 不同点：
//   - 无 SQL 依赖，所有数据存 map
//   - SKIP LOCKED 简化为"全部返回"（单实例测试足够）
//   - MarkRun 仍按 cron 表达式 + 时区计算 next_run_at
type memCronStore struct {
	mu   sync.Mutex
	jobs map[string]*CronJob
}

func newMemCronStore() *memCronStore {
	return &memCronStore{jobs: make(map[string]*CronJob)}
}

func (m *memCronStore) Get(_ context.Context, name string) (*CronJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[name]
	if !ok {
		return nil, nil
	}
	c := *j
	return &c, nil
}

func (m *memCronStore) List(_ context.Context, enabledOnly bool) ([]CronJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []CronJob
	for _, j := range m.jobs {
		if enabledOnly && !j.Enabled {
			continue
		}
		c := *j
		out = append(out, c)
	}
	return out, nil
}

func (m *memCronStore) Create(_ context.Context, j CronJob) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[j.Name]; ok {
		return 0, ErrCronJobExists
	}
	j2 := j
	m.jobs[j.Name] = &j2
	return int64(len(m.jobs)), nil
}

func (m *memCronStore) Update(_ context.Context, j CronJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[j.Name]; !ok {
		return ErrCronJobNotFound
	}
	old := m.jobs[j.Name]
	*old = j
	return nil
}

func (m *memCronStore) Delete(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[name]; !ok {
		return ErrCronJobNotFound
	}
	delete(m.jobs, name)
	return nil
}

func (m *memCronStore) Enable(_ context.Context, name string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[name]
	if !ok {
		return ErrCronJobNotFound
	}
	j.Enabled = enabled
	return nil
}

func (m *memCronStore) ClaimDue(_ context.Context, _ time.Time, limit int) ([]CronJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []CronJob
	for _, j := range m.jobs {
		if !j.Enabled {
			continue
		}
		if j.NextRunAt.IsZero() {
			continue
		}
		// 简化版：直接拿时间比，不模拟 SKIP LOCKED
		out = append(out, *j)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *memCronStore) MarkRun(_ context.Context, name, status, errMsg string, nextRunAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[name]
	if !ok {
		return ErrCronJobNotFound
	}
	j.LastRunAt = &[]time.Time{time.Now()}[0]
	j.LastStatus = status
	j.LastError = errMsg
	j.NextRunAt = nextRunAt
	j.TotalRuns++
	if status == RunStatusFailed {
		j.TotalFails++
	}
	return nil
}

// Compile-time guarantee.
var _ CronStore = (*memCronStore)(nil)

// ============================================================================
// memQueue（用于测试 scheduler 入队行为）
// ============================================================================

type memQueue struct {
	mu        sync.Mutex
	enqueued  []*Task
	failNext  bool
}

func newMemQueue() *memQueue { return &memQueue{} }

func (q *memQueue) Enqueue(_ context.Context, kind string, payload []byte, opts ...EnqueueOption) (uint64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.failNext {
		q.failNext = false
		return 0, errors.New("simulated enqueue failure")
	}
	id := uint64(len(q.enqueued) + 1)
	q.enqueued = append(q.enqueued, &Task{
		ID:          id,
		Kind:        kind,
		Payload:     payload,
		Status:      StatusPending,
		MaxAttempts: 3,
	})
	return id, nil
}

func (q *memQueue) EnqueueDelayed(_ context.Context, kind string, payload []byte, runAt time.Time) (uint64, error) {
	return q.Enqueue(context.Background(), kind, payload)
}

// 实现 Queue interface 其他方法（scheduler 仅用 Enqueue；其他用 panic 占位）
func (q *memQueue) Claim(_ context.Context, _ string, _ []string) (*Task, error) {
	return nil, nil
}
func (q *memQueue) Complete(_ context.Context, _ uint64) error       { return nil }
func (q *memQueue) Fail(_ context.Context, _ uint64, _ string, _ BackoffConfig) (bool, error) {
	return false, nil
}
func (q *memQueue) Heartbeat(_ context.Context, _ uint64) error    { return nil }
func (q *memQueue) Cancel(_ context.Context, _ uint64) error      { return nil }
func (q *memQueue) Requeue(_ context.Context, _ uint64) error     { return nil }
func (q *memQueue) ReclaimDead(_ context.Context, _ time.Time) (int, error) {
	return 0, nil
}
func (q *memQueue) Get(_ context.Context, _ uint64) (*Task, error) {
	return nil, nil
}
func (q *memQueue) List(_ context.Context, _ ListFilter) ([]*Task, int, error) {
	return nil, 0, nil
}
func (q *memQueue) Stats(_ context.Context) (Stats, error) { return Stats{}, nil }
func (q *memQueue) Cleanup(_ context.Context, _ time.Time, _ []Status) (int, error) {
	return 0, nil
}

var _ Queue = (*memQueue)(nil)

// ============================================================================
// CronScheduler.tick 单测
// ============================================================================

func TestCronScheduler_Tick_EnqueueDue(t *testing.T) {
	store := newMemCronStore()
	queue := newMemQueue()

	// 创建一个 next_run_at 已过期的 cron job
	store.jobs["cleanup"] = &CronJob{
		Name:       "cleanup",
		CronExpr:   "* * * * *",     // 每分钟
		Timezone:   "UTC",
		Kind:       "do_cleanup",
		Payload:    []byte(`{"k":"v"}`),
		Enabled:    true,
		MissPolicy: MissPolicySkip,
		NextRunAt:  time.Now().Add(-time.Minute), // 已过期
	}

	sched := NewCronScheduler(store, queue, CronSchedulerConfig{
		ScanInterval: 100 * time.Millisecond,
		BatchLimit:   10,
	})
	sched.tick(context.Background())

	// 验证：队列里有 1 条任务
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.enqueued) != 1 {
		t.Fatalf("expected 1 enqueued, got %d", len(queue.enqueued))
	}
	if queue.enqueued[0].Kind != "do_cleanup" {
		t.Errorf("kind mismatch: %s", queue.enqueued[0].Kind)
	}
	if string(queue.enqueued[0].Payload) != `{"k":"v"}` {
		t.Errorf("payload mismatch: %s", queue.enqueued[0].Payload)
	}

	// 验证：next_run_at 已推进
	store.mu.Lock()
	defer store.mu.Unlock()
	j := store.jobs["cleanup"]
	if !j.NextRunAt.After(time.Now().Add(-time.Second)) {
		t.Errorf("next_run_at not advanced: %v", j.NextRunAt)
	}
	if j.TotalRuns != 1 {
		t.Errorf("total_runs not 1: %d", j.TotalRuns)
	}
	if j.LastStatus != RunStatusSuccess {
		t.Errorf("last_status mismatch: %s", j.LastStatus)
	}
}

func TestCronScheduler_Tick_SkipsDisabled(t *testing.T) {
	store := newMemCronStore()
	queue := newMemQueue()

	// enabled=false 时不触发
	store.jobs["disabled"] = &CronJob{
		Name:       "disabled",
		CronExpr:   "* * * * *",
		Timezone:   "UTC",
		Kind:       "do_thing",
		Enabled:    false,
		NextRunAt:  time.Now().Add(-time.Minute),
	}

	NewCronScheduler(store, queue, CronSchedulerConfig{}).tick(context.Background())

	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.enqueued) != 0 {
		t.Errorf("disabled job should not enqueue, got %d", len(queue.enqueued))
	}
}

func TestCronScheduler_Tick_InvalidCronExpr(t *testing.T) {
	store := newMemCronStore()
	queue := newMemQueue()

	store.jobs["bad"] = &CronJob{
		Name:       "bad",
		CronExpr:   "not a cron",
		Timezone:   "UTC",
		Kind:       "do_thing",
		Enabled:    true,
		NextRunAt:  time.Now().Add(-time.Minute),
	}

	NewCronScheduler(store, queue, CronSchedulerConfig{}).tick(context.Background())

	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.enqueued) != 0 {
		t.Errorf("invalid cron should not enqueue")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.jobs["bad"].LastStatus != RunStatusFailed {
		t.Errorf("expected failed, got %s", store.jobs["bad"].LastStatus)
	}
}

func TestCronScheduler_Tick_InvalidTimezone(t *testing.T) {
	store := newMemCronStore()
	queue := newMemQueue()

	store.jobs["bad"] = &CronJob{
		Name:       "bad",
		CronExpr:   "* * * * *",
		Timezone:   "Invalid/Zone",
		Kind:       "do_thing",
		Enabled:    true,
		NextRunAt:  time.Now().Add(-time.Minute),
	}

	NewCronScheduler(store, queue, CronSchedulerConfig{}).tick(context.Background())

	store.mu.Lock()
	defer store.mu.Unlock()
	if store.jobs["bad"].LastStatus != RunStatusFailed {
		t.Errorf("expected failed, got %s", store.jobs["bad"].LastStatus)
	}
}

func TestCronScheduler_Tick_EnqueueFailureStillUpdatesNextRunAt(t *testing.T) {
	// 即便 Enqueue 失败，next_run_at 也必须推进——否则下次 tick 还会重复跑
	store := newMemCronStore()
	queue := newMemQueue()
	queue.failNext = true // 让下一次 Enqueue 返回 error

	store.jobs["x"] = &CronJob{
		Name:       "x",
		CronExpr:   "* * * * *",
		Timezone:   "UTC",
		Kind:       "x_kind",
		Enabled:    true,
		MissPolicy: MissPolicySkip,
		NextRunAt:  time.Now().Add(-time.Minute),
	}

	NewCronScheduler(store, queue, CronSchedulerConfig{}).tick(context.Background())

	store.mu.Lock()
	defer store.mu.Unlock()
	if !store.jobs["x"].NextRunAt.After(time.Now().Add(-time.Second)) {
		t.Error("next_run_at should advance even on enqueue failure")
	}
}

func TestCronScheduler_StartStopLifecycle(t *testing.T) {
	store := newMemCronStore()
	queue := newMemQueue()

	store.jobs["x"] = &CronJob{
		Name:       "x",
		CronExpr:   "* * * * *",
		Timezone:   "UTC",
		Kind:       "x_kind",
		Enabled:    true,
		NextRunAt:  time.Now().Add(-time.Minute),
	}

	sched := NewCronScheduler(store, queue, CronSchedulerConfig{
		ScanInterval: 50 * time.Millisecond,
		BatchLimit:   10,
	})
	if err := sched.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer sched.Stop(time.Second)

	// 等 200ms 让 tick 跑至少 1 次
	time.Sleep(200 * time.Millisecond)

	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.enqueued) < 1 {
		t.Errorf("expected ≥1 enqueue, got %d", len(queue.enqueued))
	}
}

// ============================================================================
// CronStore CRUD（用 memCronStore 验证）
// ============================================================================

func TestCronStore_CreateGetDelete(t *testing.T) {
	s := newMemCronStore()
	j := CronJob{
		Name:       "test",
		CronExpr:   "0 * * * *",
		Timezone:   "UTC",
		Kind:       "x",
		Enabled:    true,
		MissPolicy: MissPolicySkip,
		NextRunAt:  time.Now().Add(time.Hour),
	}
	if _, err := s.Create(context.Background(), j); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// duplicate → ErrCronJobExists
	if _, err := s.Create(context.Background(), j); !errors.Is(err, ErrCronJobExists) {
		t.Errorf("expected ErrCronJobExists, got %v", err)
	}

	got, err := s.Get(context.Background(), "test")
	if err != nil || got == nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Kind != "x" {
		t.Errorf("kind mismatch")
	}

	if err := s.Delete(context.Background(), "test"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ = s.Get(context.Background(), "test")
	if got != nil {
		t.Error("after Delete, Get should return nil")
	}

	// Delete not-found → ErrCronJobNotFound
	if err := s.Delete(context.Background(), "missing"); !errors.Is(err, ErrCronJobNotFound) {
		t.Errorf("expected ErrCronJobNotFound, got %v", err)
	}
}

func TestCronStore_EnableDisable(t *testing.T) {
	s := newMemCronStore()
	s.jobs["x"] = &CronJob{Name: "x", Enabled: true}

	if err := s.Enable(context.Background(), "x", false); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if s.jobs["x"].Enabled {
		t.Error("Enable false should set Enabled=false")
	}

	// not found
	if err := s.Enable(context.Background(), "nope", true); !errors.Is(err, ErrCronJobNotFound) {
		t.Errorf("expected ErrCronJobNotFound, got %v", err)
	}
}
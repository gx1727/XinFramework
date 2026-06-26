package task

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Status.IsTerminal 状态判定
// ============================================================================

func TestStatus_IsTerminal(t *testing.T) {
	cases := []struct {
		status Status
		want   bool
	}{
		{StatusPending, false},
		{StatusRunning, false},
		{StatusSucceeded, true},
		{StatusFailed, true},
		{StatusCancelled, true},
		{StatusDead, true},
		{Status("unknown"), false},
	}
	for _, c := range cases {
		t.Run(string(c.status), func(t *testing.T) {
			if got := c.status.IsTerminal(); got != c.want {
				t.Errorf("IsTerminal(%s)=%v want %v", c.status, got, c.want)
			}
		})
	}
}

// ============================================================================
// BackoffConfig.NextDelay 退避计算
// ============================================================================

func TestBackoffConfig_Exponential(t *testing.T) {
	c := BackoffConfig{
		Strategy:     BackoffExponential,
		InitialDelay: time.Second,
		MaxDelay:     time.Hour,
	}
	cases := []struct {
		attempts int
		want     time.Duration
	}{
		{1, 1 * time.Second},  // 2^0 × 1s
		{2, 2 * time.Second},  // 2^1 × 1s
		{3, 4 * time.Second},  // 2^2 × 1s
		{4, 8 * time.Second},  // 2^3 × 1s
		{5, 16 * time.Second}, // 2^4 × 1s
	}
	for _, c2 := range cases {
		if got := c.NextDelay(c2.attempts); got != c2.want {
			t.Errorf("attempts=%d: got %v want %v", c2.attempts, got, c2.want)
		}
	}
}

func TestBackoffConfig_Linear(t *testing.T) {
	c := BackoffConfig{
		Strategy:     BackoffLinear,
		InitialDelay: 30 * time.Second,
		MaxDelay:     time.Hour,
	}
	if got := c.NextDelay(1); got != 30*time.Second {
		t.Errorf("attempts=1: got %v want 30s", got)
	}
	if got := c.NextDelay(5); got != 150*time.Second {
		t.Errorf("attempts=5: got %v want 150s", got)
	}
}

func TestBackoffConfig_Fixed(t *testing.T) {
	c := BackoffConfig{
		Strategy:     BackoffFixed,
		InitialDelay: 10 * time.Second,
		MaxDelay:     time.Hour,
	}
	for _, n := range []int{1, 5, 100} {
		if got := c.NextDelay(n); got != 10*time.Second {
			t.Errorf("attempts=%d: got %v want 10s", n, got)
		}
	}
}

func TestBackoffConfig_MaxDelayTruncation(t *testing.T) {
	c := BackoffConfig{
		Strategy:     BackoffExponential,
		InitialDelay: time.Second,
		MaxDelay:     30 * time.Second,
	}
	// attempts=10: 2^9 × 1s = 512s，应被截断到 30s
	if got := c.NextDelay(10); got != 30*time.Second {
		t.Errorf("truncated: got %v want 30s", got)
	}
}

func TestBackoffConfig_DefaultsForZero(t *testing.T) {
	c := BackoffConfig{Strategy: BackoffExponential} // InitialDelay = 0
	d := c.NextDelay(1)
	if d != 30*time.Second {
		t.Errorf("default initial delay must be 30s, got %v", d)
	}
}

// ============================================================================
// HandlerFunc 适配器
// ============================================================================

func TestHandlerFunc_KindHandleTimeout(t *testing.T) {
	h := HandlerFunc{
		KindStr:  "test",
		TimeoutV: 60,
		HandleFn: func(ctx context.Context, t *Task) error { return nil },
	}
	if h.Kind() != "test" {
		t.Errorf("Kind mismatch")
	}
	if h.Timeout() != 60 {
		t.Errorf("Timeout mismatch")
	}
	if err := h.Handle(nil, nil); err != nil {
		t.Errorf("Handle should return nil, got %v", err)
	}
}

// ============================================================================
// Registry 注册表
// ============================================================================

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	h1 := HandlerFunc{KindStr: "kind_a", HandleFn: func(ctx context.Context, t *Task) error { return nil }}
	h2 := HandlerFunc{KindStr: "kind_b", HandleFn: func(ctx context.Context, t *Task) error { return nil }}

	if err := r.Register(h1); err != nil {
		t.Fatalf("Register h1: %v", err)
	}
	if err := r.Register(h2); err != nil {
		t.Fatalf("Register h2: %v", err)
	}

	got, ok := r.Get("kind_a")
	if !ok || got.Kind() != "kind_a" {
		t.Error("Get kind_a failed")
	}
	if _, ok := r.Get("nonexistent"); ok {
		t.Error("Get nonexistent should return false")
	}
}

func TestRegistry_DuplicateRegisterReturnsError(t *testing.T) {
	r := NewRegistry()
	h := HandlerFunc{KindStr: "kind_a", HandleFn: func(ctx context.Context, t *Task) error { return nil }}
	if err := r.Register(h); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := r.Register(h)
	if err != ErrHandlerAlreadyExists {
		t.Errorf("expected ErrHandlerAlreadyExists, got %v", err)
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	if len(r.Names()) != 0 {
		t.Errorf("empty registry should have no names, got %v", r.Names())
	}
	r.Register(HandlerFunc{KindStr: "kind_a", HandleFn: func(ctx context.Context, t *Task) error { return nil }})
	r.Register(HandlerFunc{KindStr: "kind_b", HandleFn: func(ctx context.Context, t *Task) error { return nil }})
	names := r.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %v", names)
	}
}

func TestLookupHandler_NotFound(t *testing.T) {
	// LookupHandler 走全局 registry，所以测试需在隔离环境：直接测 Get。
	r := NewRegistry()
	_, err := LookupHandler("nonexistent_global_kind")
	if err != ErrHandlerNotFound {
		t.Errorf("expected ErrHandlerNotFound, got %v", err)
	}
	// 注意：上面这条 case 实际查的是全局 registry，多个测试并行可能互相影响。
	// 真正可靠的方式是直接用 r.Get。这里仅验证错误类型。
	_ = r
}

// ============================================================================
// EnqueueOption 默认值
// ============================================================================

func TestEnqueueOption_Defaults(t *testing.T) {
	cfg := defaultEnqueueConfig()
	if cfg.maxAttempts != 3 {
		t.Errorf("default max_attempts must be 3, got %d", cfg.maxAttempts)
	}
	if cfg.timeoutSec != 300 {
		t.Errorf("default timeout_sec must be 300, got %d", cfg.timeoutSec)
	}
	if cfg.priority != 0 {
		t.Errorf("default priority must be 0, got %d", cfg.priority)
	}
}

// ============================================================================
// MarshalPayload
// ============================================================================

func TestMarshalPayload(t *testing.T) {
	type S struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	data := MarshalPayload(S{Name: "test", Count: 42})
	if string(data) != `{"name":"test","count":42}` {
		t.Errorf("unexpected payload: %s", data)
	}
}
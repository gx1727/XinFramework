package login_security

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestLogNotifier_SatisfiesInterface 编译期 + 运行期双重断言。
func TestLogNotifier_SatisfiesInterface(t *testing.T) {
	var _ Notifier = NewLogNotifier()
}

// TestLogNotifier_NotifyDoesNotError 验证 LogNotifier.Notify 不返回 error。
// 即便没有真实短信/邮件通道，写日志也算"通知成功"。
func TestLogNotifier_NotifyDoesNotError(t *testing.T) {
	n := NewLogNotifier()
	err := n.Notify(context.Background(), NotificationPayload{
		Channel:   ChannelEmail,
		Recipient: "test@example.com",
		Subject:   "test",
		Body:      "hello",
		AccountID: 1,
		Reason:    "unit_test",
	})
	if err != nil {
		t.Errorf("LogNotifier.Notify must not error: %v", err)
	}
}

// TestMultiNotifier_EmptyListNoop 验证 notifier 列表为空时不返回错误。
func TestMultiNotifier_EmptyListNoop(t *testing.T) {
	m := NewMultiNotifier()
	err := m.Notify(context.Background(), NotificationPayload{
		Channel: ChannelSMS, Body: "x",
	})
	if err != nil {
		t.Errorf("empty MultiNotifier must not error: %v", err)
	}
}

// TestMultiNotifier_AtLeastOneSuccess 即使有 notifier 报错，只要一个成功即可。
func TestMultiNotifier_AtLeastOneSuccess(t *testing.T) {
	failNotifier := &fakeNotifier{err: errors.New("boom")}
	okNotifier := &fakeNotifier{}

	m := NewMultiNotifier(failNotifier, okNotifier)
	err := m.Notify(context.Background(), NotificationPayload{Body: "x"})
	if err != nil {
		t.Errorf("MultiNotifier should return nil if at least one notifier succeeds, got %v", err)
	}
	if !okNotifier.called {
		t.Error("okNotifier must be invoked")
	}
	if !failNotifier.called {
		t.Error("failNotifier must be invoked")
	}
}

// TestMultiNotifier_AllFailed_ReturnsLastError 全失败时返回最后一个 error。
func TestMultiNotifier_AllFailed_ReturnsLastError(t *testing.T) {
	n1 := &fakeNotifier{err: errors.New("e1")}
	n2 := &fakeNotifier{err: errors.New("e2")}
	m := NewMultiNotifier(n1, n2)
	err := m.Notify(context.Background(), NotificationPayload{Body: "x"})
	if err == nil || err.Error() != "e2" {
		t.Errorf("expected last error e2, got %v", err)
	}
}

// fakeNotifier 是测试用 Notifier stub：返回预设 error。
type fakeNotifier struct {
	called bool
	err    error
}

func (f *fakeNotifier) Notify(_ context.Context, _ NotificationPayload) error {
	f.called = true
	return f.err
}

// ============================================================================
// AccountLock 行为
// ============================================================================

// TestAccountLock_IsActive 验证锁定记录在 locked_until 之前算 active。
func TestAccountLock_IsActive(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name  string
		lock  *AccountLock
		now   time.Time
		want  bool
	}{
		{"nil-lock", nil, now, false},
		{"active", &AccountLock{LockedUntil: now.Add(time.Minute)}, now, true},
		{"expired", &AccountLock{LockedUntil: now.Add(-time.Minute)}, now, false},
		{"exactly-now", &AccountLock{LockedUntil: now}, now, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.lock.IsActive(c.now)
			if got != c.want {
				t.Errorf("IsActive(%v)=%v want %v", c.lock, got, c.want)
			}
		})
	}
}

// ============================================================================
// detectAnomaly（异地判定核心）
// ============================================================================

func TestDetectAnomaly_NoHistory(t *testing.T) {
	now := LoginHistoryEntry{IP: "1.1.1.1", DeviceID: "d1"}
	sig := detectAnomaly(now, nil, false)
	if sig.IsAnomaly {
		t.Error("first login must not be anomaly")
	}
}

func TestDetectAnomaly_NewIP(t *testing.T) {
	now := LoginHistoryEntry{IP: "2.2.2.2"}
	recent := []LoginHistoryEntry{
		{IP: "1.1.1.1"},
		{IP: "1.1.1.1"},
		{IP: "1.1.1.2"},
	}
	sig := detectAnomaly(now, recent, false)
	if !sig.IsAnomaly {
		t.Error("new IP must trigger anomaly")
	}
	found := false
	for _, r := range sig.Reasons {
		if r == "new_ip" {
			found = true
		}
	}
	if !found {
		t.Errorf("reason 'new_ip' not found: %v", sig.Reasons)
	}
}

func TestDetectAnomaly_SameIP_NoAnomaly(t *testing.T) {
	now := LoginHistoryEntry{IP: "1.1.1.1"}
	recent := []LoginHistoryEntry{
		{IP: "1.1.1.1"},
		{IP: "1.1.1.2"},
		{IP: "1.1.1.1"},
	}
	sig := detectAnomaly(now, recent, false)
	if sig.IsAnomaly {
		t.Errorf("known IP must not trigger anomaly, got reasons=%v", sig.Reasons)
	}
}

func TestDetectAnomaly_DeviceMatch_Disabled(t *testing.T) {
	// AnomalyDeviceMatch=false → device_id 不参与判定
	now := LoginHistoryEntry{IP: "1.1.1.1", DeviceID: "newdev"}
	recent := []LoginHistoryEntry{
		{IP: "1.1.1.1", DeviceID: "olddev"},
	}
	sig := detectAnomaly(now, recent, false)
	if sig.IsAnomaly {
		t.Errorf("device_match=false must not flag device change, got reasons=%v", sig.Reasons)
	}
}

func TestDetectAnomaly_DeviceMatch_Enabled(t *testing.T) {
	// AnomalyDeviceMatch=true → 新 device_id 也算"异地"
	now := LoginHistoryEntry{IP: "1.1.1.1", DeviceID: "newdev"}
	recent := []LoginHistoryEntry{
		{IP: "1.1.1.1", DeviceID: "olddev"},
	}
	sig := detectAnomaly(now, recent, true)
	if !sig.IsAnomaly {
		t.Errorf("new device must trigger anomaly, got reasons=%v", sig.Reasons)
	}
}

func TestDetectAnomaly_MultipleIPs_KnownIPsList(t *testing.T) {
	now := LoginHistoryEntry{IP: "9.9.9.9"}
	recent := []LoginHistoryEntry{
		{IP: "1.1.1.1"},
		{IP: "1.1.1.2"},
		{IP: "1.1.1.1"},
	}
	sig := detectAnomaly(now, recent, false)
	if len(sig.KnownIPs) != 2 {
		t.Errorf("expected 2 known IPs, got %v", sig.KnownIPs)
	}
}

// ============================================================================
// SecurityService.CheckLock（in-memory 锁定管理）
// ============================================================================

// memLockManager 是 LockManager 的内存实现，仅用于单元测试。
// 不做持久化，但记录所有操作便于断言。
type memLockManager struct {
	locks map[string]*AccountLock
}

func newMemLockManager() *memLockManager {
	return &memLockManager{locks: make(map[string]*AccountLock)}
}

func (m *memLockManager) Get(_ context.Context, account string) (*AccountLock, error) {
	return m.locks[account], nil
}

func (m *memLockManager) Lock(_ context.Context, account string, duration time.Duration, reason LockReason, attempts int, ip string) error {
	m.locks[account] = &AccountLock{
		Account:     account,
		LockedUntil: time.Now().Add(duration),
		Reason:      reason,
		Attempts:    attempts,
		IP:          ip,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (m *memLockManager) Unlock(_ context.Context, account string) error {
	if _, ok := m.locks[account]; !ok {
		return ErrLockNotFound
	}
	delete(m.locks, account)
	return nil
}

func (m *memLockManager) CleanupExpired(_ context.Context, now time.Time) (int, error) {
	n := 0
	for k, v := range m.locks {
		if !v.IsActive(now) {
			delete(m.locks, k)
			n++
		}
	}
	return n, nil
}

func TestSecurityService_CheckLock_NoLock(t *testing.T) {
	svc := NewSecurityService(DefaultSecurityConfig(), newMemLockManager(), nil, nil, nil, nil)
	lock, err := svc.CheckLock(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lock != nil {
		t.Errorf("expected nil lock, got %+v", lock)
	}
}

func TestSecurityService_CheckLock_AfterLock(t *testing.T) {
	lm := newMemLockManager()
	svc := NewSecurityService(DefaultSecurityConfig(), lm, nil, nil, nil, nil)
	ctx := context.Background()

	if err := lm.Lock(ctx, "alice", 30*time.Minute, LockTooManyFailures, 5, "1.1.1.1"); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	lock, err := svc.CheckLock(ctx, "alice")
	if err != nil {
		t.Fatalf("CheckLock: %v", err)
	}
	if lock == nil {
		t.Fatal("expected lock, got nil")
	}
	if lock.Reason != LockTooManyFailures {
		t.Errorf("reason mismatch: %v", lock.Reason)
	}
	if lock.Attempts != 5 {
		t.Errorf("attempts mismatch: %v", lock.Attempts)
	}
}

func TestSecurityService_Disabled_NoOps(t *testing.T) {
	// Enabled=false 时所有方法应跳过
	cfg := SecurityConfig{Enabled: false}
	svc := NewSecurityService(cfg, newMemLockManager(), nil, nil, nil, nil)
	ctx := context.Background()

	lock, _ := svc.CheckLock(ctx, "alice")
	if lock != nil {
		t.Error("disabled service must return nil lock")
	}
	count, triggered, _ := svc.RecordFailure(ctx, "alice", "1.1.1.1", "ua", FailureInvalidPassword, ScopeTenant, 1)
	if count != 0 || triggered {
		t.Errorf("disabled service must not record failures: count=%d triggered=%v", count, triggered)
	}
}
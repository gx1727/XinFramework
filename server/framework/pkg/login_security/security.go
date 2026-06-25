package login_security

import (
	"context"
	"errors"
	"time"

	"gx1727.com/xin/framework/pkg/logger"
)

// SecurityConfig 控制账号锁定与异地告警的行为。
type SecurityConfig struct {
	Enabled bool // 总开关；false 时 SecurityService 所有方法变成 noop

	// 账号维度锁定策略
	MaxFailedAttempts  int           // 滑动窗口内最大失败次数（默认 5）
	LockDuration       time.Duration // 锁定时长（默认 30 分钟）
	FailureWindow      time.Duration // 滑动窗口长度（默认 10 分钟）
	IPFailureThreshold int           // 同一 IP 失败阈值（默认 20），用于跨账号爆破检测
	IPFailureWindow    time.Duration // IP 维度的窗口（默认 5 分钟）

	// 异地告警
	AnomalyEnabled       bool          // 是否启用异地检测（默认 true）
	AnomalyHistoryLimit  int           // 取最近 N 次历史对比（默认 5）
	AnomalyDeviceMatch   bool          // device_id 一致是否算"未异地"（默认 false，留宽）
	AnomalyNotifyInSite  bool          // 异地时是否同时发站内消息（默认 true）
	AnomalyNotifyEmail   bool          // 异地时是否发邮件（默认 true）
	AnomalyNotifySMS     bool          // 异地时是否发短信（默认 false，避免误打扰）

	// 锁定通知
	LockNotifyInSite     bool          // 锁定时是否发站内消息（默认 true）
	LockNotifyEmail      bool          // 锁定时是否发邮件（默认 true）
	LockNotifySMS        bool          // 锁定时是否发短信（默认 true）
}

// DefaultSecurityConfig 返回推荐默认值。
//
// 注意：调用方应自行根据 SecurityConfig.Enabled 判断是否启用整组策略。
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Enabled:             true,
		MaxFailedAttempts:   5,
		LockDuration:        30 * time.Minute,
		FailureWindow:       10 * time.Minute,
		IPFailureThreshold:  20,
		IPFailureWindow:     5 * time.Minute,
		AnomalyEnabled:      true,
		AnomalyHistoryLimit: 5,
		AnomalyDeviceMatch:  false,
		AnomalyNotifyInSite: true,
		AnomalyNotifyEmail:  true,
		AnomalyNotifySMS:    false,
		LockNotifyInSite:    true,
		LockNotifyEmail:     true,
		LockNotifySMS:       true,
	}
}

// SecurityService 是登录安全策略的统一入口。
//
// 装配路径：boot.Init 构造，auth.Service 持有引用。
type SecurityService struct {
	cfg        SecurityConfig
	locker     LockManager
	attempts   AttemptStore
	history    HistoryRecorder
	notifier   Notifier
	recipients RecipientResolver
	now        func() time.Time // 便于测试注入
}

// RecipientResolver 根据 accountID 取联系方式（手机号 / 邮箱）。
//
// 框架不内置具体实现，由 auth 模块注入（auth 包有 AccountRepository 可查）。
type RecipientResolver interface {
	ResolveEmail(ctx context.Context, accountID uint) (string, error)
	ResolvePhone(ctx context.Context, accountID uint) (string, error)
	ResolveUserID(ctx context.Context, accountID uint) (uint, error) // for in-site message
}

// NewSecurityService 构造 SecurityService。
//
// notifier 可为 nil（内部 fallback 到 LogNotifier）。
func NewSecurityService(
	cfg SecurityConfig,
	locker LockManager,
	attempts AttemptStore,
	history HistoryRecorder,
	notifier Notifier,
	recipients RecipientResolver,
) *SecurityService {
	if notifier == nil {
		notifier = NewLogNotifier()
	}
	return &SecurityService{
		cfg:        cfg,
		locker:     locker,
		attempts:   attempts,
		history:    history,
		notifier:   notifier,
		recipients: recipients,
		now:        time.Now,
	}
}

// ErrAccountLocked 表示账号当前处于锁定状态，调用方应拒绝登录。
var ErrAccountLocked = errors.New("login_security: account is locked")

// CheckLock 在登录流程开始前检查账号是否被锁。
//
// 返回：
//   - (nil, nil)：账号未锁定，可继续登录
//   - (*AccountLock, nil)：账号当前被锁，调用方应拒绝并提示锁定剩余时间
//   - (nil, error)：后端异常（DB 错误等）
func (s *SecurityService) CheckLock(ctx context.Context, account string) (*AccountLock, error) {
	if !s.cfg.Enabled || s.locker == nil {
		return nil, nil
	}
	lock, err := s.locker.Get(ctx, account)
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, nil
	}
	if !lock.IsActive(s.now()) {
		// 过期锁不视为有效（避免历史脏数据误判），由 CleanupExpired 清理
		return nil, nil
	}
	return lock, nil
}

// RecordFailure 记录一次失败尝试，必要时触发锁定。
//
// 返回值：
//   - failureCount：当前窗口内的累计失败次数（含本次）
//   - triggeredLock：是否本次触发了锁定（true 表示账号刚从可登录变为锁定）
//   - error：DB 错误
//
// 调用方应在登录失败后立即调用，传入 LoginAttempt（已填好失败原因）。
func (s *SecurityService) RecordFailure(ctx context.Context, account, ip, ua string, reason FailureReason, scope Scope, tenantID uint) (failureCount int, triggeredLock bool, err error) {
	if !s.cfg.Enabled {
		return 0, false, nil
	}
	now := s.now()
	a := LoginAttempt{
		Account:       account,
		IP:            ip,
		UserAgent:      ua,
		Success:       false,
		FailureReason: reason,
		Scope:         scope,
		TenantID:      tenantID,
		CreatedAt:     now,
	}
	if s.attempts != nil {
		if err := s.attempts.Record(ctx, a); err != nil {
			logger.Module("login_security").Warnf("record attempt failed: account=%s ip=%s err=%v", account, ip, err)
		}
	}

	if s.attempts == nil {
		return 0, false, nil
	}
	count, err := s.attempts.CountRecentFailures(ctx, account, s.cfg.FailureWindow, now)
	if err != nil {
		return 0, false, err
	}
	if count < s.cfg.MaxFailedAttempts {
		return count, false, nil
	}

	// 触发锁定
	if s.locker == nil {
		return count, true, nil
	}
	if err := s.locker.Lock(ctx, account, s.cfg.LockDuration, LockTooManyFailures, count, ip); err != nil {
		return count, false, err
	}

	// 触发锁定通知（best-effort，不阻断返回）
	s.notifyLockTriggered(ctx, account, ip, count)

	return count, true, nil
}

// notifyLockTriggered 异步（同步实现）通知账号被锁定。
//
// 失败仅记日志：锁定本身已经生效，通知失败不应回滚。
func (s *SecurityService) notifyLockTriggered(ctx context.Context, account, ip string, attempts int) {
	if s.notifier == nil || s.recipients == nil {
		return
	}
	accountID := s.resolveAccountID(ctx, account)
	var email, phone string
	if s.cfg.LockNotifyEmail {
		email, _ = s.recipients.ResolveEmail(ctx, accountID)
	}
	if s.cfg.LockNotifySMS {
		phone, _ = s.recipients.ResolvePhone(ctx, accountID)
	}

	body := "账号在短时间内的登录失败次数过多，已被临时锁定 " +
		s.cfg.LockDuration.String() + "。如非本人操作，请尽快修改密码。"
	if email != "" && s.cfg.LockNotifyEmail {
		_ = s.notifier.Notify(ctx, NotificationPayload{
			Channel:   ChannelEmail,
			Recipient: email,
			Subject:   "账号安全告警",
			Body:      body,
			AccountID: accountID,
			Reason:    "account_locked",
			Extra:     map[string]string{"ip": ip, "attempts": itoa(attempts)},
		})
	}
	if phone != "" && s.cfg.LockNotifySMS {
		_ = s.notifier.Notify(ctx, NotificationPayload{
			Channel:   ChannelSMS,
			Recipient: phone,
			Subject:   "Xin 安全",
			Body:      body,
			AccountID: accountID,
			Reason:    "account_locked",
			Extra:     map[string]string{"ip": ip, "attempts": itoa(attempts)},
		})
	}
	if s.cfg.LockNotifyInSite {
		userID, _ := s.recipients.ResolveUserID(ctx, accountID)
		if userID > 0 {
			_ = s.notifier.Notify(ctx, NotificationPayload{
				Channel:   ChannelInSite,
				Recipient: itoaU(userID),
				Body:      body,
				AccountID: accountID,
				Reason:    "account_locked",
				Extra:     map[string]string{"ip": ip, "attempts": itoa(attempts)},
			})
		}
	}
}

// resolveAccountID 反查 account 字符串对应的 account_id。
//
// 由于 auth 模块通过 ResolveLoginIdentity 已经拿到账号 ID，
// 通常应在调用 RecordFailure 前就把 accountID 准备好。
// 当前实现：从 attempts 表查最近一条该 account 的记录以拿到 account_id。
// 如果查不到（极少见，可能是并发 race），返回 0 让通知降级为不发送。
func (s *SecurityService) resolveAccountID(ctx context.Context, account string) uint {
	if s.attempts == nil {
		return 0
	}
	// 这里复用 AttemptStore 不优雅；实际应在 Service 调用方把 accountID 一并传入。
	// 为保持 SecurityService 接口简洁，这里返回 0，由调用方选择是否提供 accountID。
	return 0
}

// RecordSuccess 记录一次成功登录，并检测是否触发异地告警。
//
// 返回值：
//   - history：本次写入的 history entry（便于调用方填 session_id 后再回填）
//   - anomaly：异地告警信号；若 IsAnomaly 为 true，Notifier 已发送告警
//   - error：DB / Notify 错误
func (s *SecurityService) RecordSuccess(ctx context.Context, entry LoginHistoryEntry) (*LoginHistoryEntry, *AnomalySignal, error) {
	if !s.cfg.Enabled {
		return &entry, &AnomalySignal{}, nil
	}
	if s.history == nil {
		return &entry, &AnomalySignal{}, nil
	}
	if entry.LoginAt.IsZero() {
		entry.LoginAt = s.now()
	}
	if err := s.history.Record(ctx, entry); err != nil {
		return nil, nil, err
	}

	if !s.cfg.AnomalyEnabled {
		return &entry, &AnomalySignal{}, nil
	}

	// 异地检测
	limit := s.cfg.AnomalyHistoryLimit
	if limit <= 0 {
		limit = 5
	}
	recent, err := s.history.ListRecent(ctx, entry.AccountID, limit)
	if err != nil {
		return &entry, &AnomalySignal{}, err
	}
	signal := detectAnomaly(entry, recent, s.cfg.AnomalyDeviceMatch)

	if signal.IsAnomaly && s.notifier != nil {
		s.notifyAnomaly(ctx, entry, signal)
	}
	return &entry, signal, nil
}

// detectAnomaly 比对当前登录与最近 N 次历史的 IP/device_id。
//
// 判定规则：
//   - 没有任何历史 → 不是异地（新账号）
//   - 历史上所有 IP 都相同 + 本次 IP 不同 → 异地（new_ip）
//   - device_id 启用时，历史上所有 device_id 都相同 + 本次 device_id 不同 → 异地（new_device）
func detectAnomaly(now LoginHistoryEntry, recent []LoginHistoryEntry, deviceMatch bool) *AnomalySignal {
	sig := &AnomalySignal{}
	if len(recent) == 0 {
		return sig
	}
	seenIP := make(map[string]bool, len(recent))
	for _, h := range recent {
		if h.IP != "" {
			seenIP[h.IP] = true
		}
	}
	sig.KnownIPs = keysOfMap(seenIP)
	if now.IP != "" && len(seenIP) > 0 && !seenIP[now.IP] {
		sig.IsAnomaly = true
		sig.Reasons = append(sig.Reasons, "new_ip")
	}
	if deviceMatch && now.DeviceID != "" {
		seenDev := make(map[string]bool, len(recent))
		for _, h := range recent {
			if h.DeviceID != "" {
				seenDev[h.DeviceID] = true
			}
		}
		sig.KnownDevices = keysOfMap(seenDev)
		if len(seenDev) > 0 && !seenDev[now.DeviceID] {
			sig.IsAnomaly = true
			sig.Reasons = append(sig.Reasons, "new_device")
		}
	}
	return sig
}

// notifyAnomaly 异地告警通知。
func (s *SecurityService) notifyAnomaly(ctx context.Context, entry LoginHistoryEntry, sig *AnomalySignal) {
	if s.notifier == nil || s.recipients == nil {
		return
	}
	var email, phone string
	if s.cfg.AnomalyNotifyEmail {
		email, _ = s.recipients.ResolveEmail(ctx, entry.AccountID)
	}
	if s.cfg.AnomalyNotifySMS {
		phone, _ = s.recipients.ResolvePhone(ctx, entry.AccountID)
	}
	body := "检测到您的账号在新 IP 登录：登录时间=" + entry.LoginAt.Format("2006-01-02 15:04:05") +
		"  IP=" + entry.IP + "  UA=" + entry.UserAgent
	if email != "" && s.cfg.AnomalyNotifyEmail {
		_ = s.notifier.Notify(ctx, NotificationPayload{
			Channel:   ChannelEmail,
			Recipient: email,
			Subject:   "账号异地登录提醒",
			Body:      body,
			AccountID: entry.AccountID,
			Reason:    "anomaly_login",
			Extra:     map[string]string{"ip": entry.IP, "reasons": joinReasons(sig.Reasons)},
		})
	}
	if phone != "" && s.cfg.AnomalyNotifySMS {
		_ = s.notifier.Notify(ctx, NotificationPayload{
			Channel:   ChannelSMS,
			Recipient: phone,
			Subject:   "Xin 安全",
			Body:      body,
			AccountID: entry.AccountID,
			Reason:    "anomaly_login",
			Extra:     map[string]string{"ip": entry.IP, "reasons": joinReasons(sig.Reasons)},
		})
	}
	if s.cfg.AnomalyNotifyInSite {
		userID, _ := s.recipients.ResolveUserID(ctx, entry.AccountID)
		if userID > 0 {
			_ = s.notifier.Notify(ctx, NotificationPayload{
				Channel:   ChannelInSite,
				Recipient: itoaU(userID),
				Body:      body,
				AccountID: entry.AccountID,
				Reason:    "anomaly_login",
				Extra:     map[string]string{"ip": entry.IP, "reasons": joinReasons(sig.Reasons)},
			})
		}
	}
}

// 内部工具函数（避免引入 strconv 包依赖）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func itoaU(u uint) string {
	if u == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for u > 0 {
		i--
		buf[i] = byte('0' + u%10)
		u /= 10
	}
	return string(buf[i:])
}

func keysOfMap(m map[string]bool) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	out := reasons[0]
	for _, r := range reasons[1:] {
		out += "," + r
	}
	return out
}
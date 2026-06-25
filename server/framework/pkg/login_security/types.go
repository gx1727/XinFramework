// Package login_security 实现账号登录安全策略：失败计数、账号锁定、
// 登录历史、IP 审计、异地告警。
//
// 核心组件：
//   - SecurityService：组合 LockManager / AttemptStore / HistoryRecorder / Notifier
//   - LockManager：账号维度的锁定/解锁/状态查询
//   - AttemptStore：登录尝试流水（含成功与失败），用于滑动窗口判定
//   - HistoryRecorder：仅记录成功登录的 IP/UA/位置，用于异地告警
//   - Notifier：通知通道抽象（短信 / 邮件 / 站内消息），默认 LogNotifier 仅写日志
//
// 调用方：apps/boot/auth 的 Service 在 Login / LoginPrecheck / PlatformLogin 流程中
// 调用 SecurityService.CheckLock / RecordAttempt / RecordSuccess 即可。
package login_security

import (
	"context"
	"time"
)

// FailureReason 登录失败的语义化原因，用于 lock_reason / failure_reason 列。
type FailureReason string

const (
	FailureInvalidPassword  FailureReason = "invalid_password"
	FailureAccountNotFound  FailureReason = "account_not_found"
	FailureUserDisabled     FailureReason = "user_disabled"
	FailureAccountLocked    FailureReason = "locked"
	FailureTenantNotFound   FailureReason = "tenant_not_found"
	FailureNoLoginIdentity  FailureReason = "no_login_identity"
)

// LockReason 账号锁定原因。
type LockReason string

const (
	LockTooManyFailures LockReason = "too_many_failures"
	LockManual          LockReason = "manual"
	LockSecurityAlert   LockReason = "security_alert"
)

// Scope 登录作用域：与 auth.LoginScope 保持一致但独立定义避免循环依赖。
type Scope string

const (
	ScopeTenant   Scope = "tenant"
	ScopePlatform Scope = "platform"
	ScopePrecheck Scope = "precheck"
)

// LoginAttempt 一次登录尝试的完整记录。
type LoginAttempt struct {
	Account       string
	IP            string
	UserAgent     string
	Success       bool
	FailureReason FailureReason
	Scope         Scope
	TenantID      uint
	CreatedAt     time.Time
}

// AccountLock 当前生效的账号锁定记录。
type AccountLock struct {
	Account      string
	LockedUntil  time.Time
	Reason       LockReason
	Attempts     int
	IP           string
	CreatedAt    time.Time
}

// IsActive 报告该锁定记录在当前时刻是否仍生效。
// 框架使用前应同时校验 LockedUntil 是否过期——过期记录由 CleanupExpired 清理。
func (l *AccountLock) IsActive(now time.Time) bool {
	return l != nil && now.Before(l.LockedUntil)
}

// LoginHistoryEntry 一次成功登录的完整记录。
type LoginHistoryEntry struct {
	AccountID uint
	UserID    uint
	TenantID  uint
	Scope     Scope
	IP        string
	UserAgent string
	DeviceID  string
	Location  string
	SessionID string
	LoginAt   time.Time
}

// AnomalySignal 异地登录信号。SecurityService 在 RecordSuccess 时检测并返回。
type AnomalySignal struct {
	IsAnomaly     bool     // 是否判定为"异地"
	Reasons       []string // 命中规则：new_ip / new_device / new_location
	KnownIPs      []string // 该账号最近 N 次登录的 IP 列表（取证）
	KnownDevices  []string // 该账号最近 N 次登录的 device_id 列表
}

// ContextKey 注入到 ctx 的私有 key（避免与其他包冲突）。
type contextKey struct{}

// AttemptRecordFromContext 取 ctx 里的 LoginAttempt（handler 中间件注入）。
// 用于在 auth handler 中调用 SecurityService 时复用请求级元数据。
func AttemptRecordFromContext(ctx context.Context) *LoginAttempt {
	v, _ := ctx.Value(contextKey{}).(*LoginAttempt)
	return v
}

// WithAttemptRecord 把 attempt 注入 ctx，便于跨方法传递。
func WithAttemptRecord(ctx context.Context, a *LoginAttempt) context.Context {
	return context.WithValue(ctx, contextKey{}, a)
}
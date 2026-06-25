package login_security

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AttemptStore 负责登录尝试的写入与按滑动窗口查询。
//
// 设计目的：避免每次登录都 COUNT(*) 全表——只查"最近 N 分钟"的失败记录。
// 通过 (account, created_at DESC) 索引保证窗口查询是 O(失败次数)。
type AttemptStore interface {
	// Record 写入一次尝试（含成功与失败）。
	Record(ctx context.Context, attempt LoginAttempt) error

	// CountRecentFailures 统计账号在 [now - window, now] 区间内的失败次数。
	// 用于触发锁定的滑动窗口判定。
	CountRecentFailures(ctx context.Context, account string, window time.Duration, now time.Time) (int, error)

	// CountRecentFailuresByIP 统计 IP 在 [now - window, now] 区间内的失败次数。
	// 用于防止"同一 IP 跨账号爆破"。
	CountRecentFailuresByIP(ctx context.Context, ip string, window time.Duration, now time.Time) (int, error)
}

// PGAttemptStore 是 AttemptStore 的 PostgreSQL 实现。
type PGAttemptStore struct {
	pool *pgxpool.Pool
}

// NewPGAttemptStore 构造基于 pgxpool 的 AttemptStore。
func NewPGAttemptStore(pool *pgxpool.Pool) *PGAttemptStore {
	return &PGAttemptStore{pool: pool}
}

// Record 写入一条 LoginAttempt。
//
// success=true 时 failure_reason 列写 NULL（PG 语义：失败原因只在失败时记录）。
func (s *PGAttemptStore) Record(ctx context.Context, a LoginAttempt) error {
	if s.pool == nil {
		return errors.New("login_security: pg pool is nil")
	}
	createdAt := a.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	scope := string(a.Scope)
	if a.Scope == "" {
		scope = string(ScopeTenant)
	}
	var tenantID any
	if a.TenantID > 0 {
		tenantID = a.TenantID
	}
	var reason any
	if a.FailureReason != "" {
		reason = string(a.FailureReason)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO login_attempts
			(account, ip, user_agent, success, failure_reason, scope, tenant_id, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8)
	`, a.Account, a.IP, a.UserAgent, a.Success, reason, scope, tenantID, createdAt)
	return err
}

// CountRecentFailures 实现 AttemptStore.CountRecentFailures。
//
// 使用 idx_login_attempts_account_time 索引，查询窗口限定在 [now - window, now]。
func (s *PGAttemptStore) CountRecentFailures(ctx context.Context, account string, window time.Duration, now time.Time) (int, error) {
	if s.pool == nil {
		return 0, errors.New("login_security: pg pool is nil")
	}
	if account == "" {
		return 0, nil
	}
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM login_attempts
		WHERE account = $1
		  AND success = FALSE
		  AND created_at >= $2
	`, account, now.Add(-window)).Scan(&n)
	return n, err
}

// CountRecentFailuresByIP 实现 AttemptStore.CountRecentFailuresByIP。
//
// 使用 idx_login_attempts_ip_time 索引。
func (s *PGAttemptStore) CountRecentFailuresByIP(ctx context.Context, ip string, window time.Duration, now time.Time) (int, error) {
	if s.pool == nil {
		return 0, errors.New("login_security: pg pool is nil")
	}
	if ip == "" {
		return 0, nil
	}
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM login_attempts
		WHERE ip = $1
		  AND success = FALSE
		  AND created_at >= $2
	`, ip, now.Add(-window)).Scan(&n)
	return n, err
}

// Compile-time guarantee.
var _ AttemptStore = (*PGAttemptStore)(nil)
package login_security

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrLockNotFound 在 LockManager.Get / LockManager.Unlock 查不到时返回。
var ErrLockNotFound = errors.New("login_security: lock not found")

// LockManager 负责账号锁定的查询 / 设置 / 解除 / 清理。
//
// 底层存储是 account_locks 表。接口被设计成可在测试中替换为内存实现
// （不需要真实 PG）。
type LockManager interface {
	// Get 查询账号当前的锁定记录。若未锁定或已过期返回 (nil, nil)。
	Get(ctx context.Context, account string) (*AccountLock, error)

	// Lock 设置账号锁定，locked_until = now + duration。
	// 若已存在锁定记录，覆盖 locked_until（保留最早的 CreatedAt 便于审计）。
	Lock(ctx context.Context, account string, duration time.Duration, reason LockReason, attempts int, ip string) error

	// Unlock 解除账号锁定。不存在时返回 ErrLockNotFound。
	Unlock(ctx context.Context, account string) error

	// CleanupExpired 删除所有 locked_until <= now 的记录。建议定时任务每天跑一次。
	CleanupExpired(ctx context.Context, now time.Time) (int, error)
}

// PGLockManager 是 LockManager 的 PostgreSQL 实现。
type PGLockManager struct {
	pool *pgxpool.Pool
}

// NewPGLockManager 构造基于 pgxpool 的 LockManager。
func NewPGLockManager(pool *pgxpool.Pool) *PGLockManager {
	return &PGLockManager{pool: pool}
}

// Get 实现 LockManager.Get。
func (m *PGLockManager) Get(ctx context.Context, account string) (*AccountLock, error) {
	if account == "" {
		return nil, nil
	}
	if m.pool == nil {
		return nil, errors.New("login_security: pg pool is nil")
	}
	row := m.pool.QueryRow(ctx, `
		SELECT account, locked_until, reason, attempts, COALESCE(ip, ''), created_at
		FROM account_locks
		WHERE account = $1
		LIMIT 1`, account)

	var lock AccountLock
	var reason string
	if err := row.Scan(&lock.Account, &lock.LockedUntil, &reason, &lock.Attempts, &lock.IP, &lock.CreatedAt); err != nil {
		// pgx.ErrNoRows 时返回 (nil, nil) 表示未锁定
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	lock.Reason = LockReason(reason)
	return &lock, nil
}

// Lock 实现 LockManager.Lock。
//
// 使用 UPSERT：账号已锁定时刷新 locked_until，保留 reason 与 attempts。
// 这样后续失败次数不会"重置"，便于审计累计失败总数。
func (m *PGLockManager) Lock(ctx context.Context, account string, duration time.Duration, reason LockReason, attempts int, ip string) error {
	if account == "" {
		return errors.New("login_security: empty account")
	}
	if m.pool == nil {
		return errors.New("login_security: pg pool is nil")
	}
	lockedUntil := time.Now().Add(duration)
	_, err := m.pool.Exec(ctx, `
		INSERT INTO account_locks (account, locked_until, reason, attempts, ip, created_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NOW())
		ON CONFLICT (account) DO UPDATE SET
			locked_until = EXCLUDED.locked_until,
			reason = EXCLUDED.reason,
			attempts = EXCLUDED.attempts,
			ip = EXCLUDED.ip
	`, account, lockedUntil, string(reason), attempts, ip)
	return err
}

// Unlock 实现 LockManager.Unlock。
func (m *PGLockManager) Unlock(ctx context.Context, account string) error {
	if m.pool == nil {
		return errors.New("login_security: pg pool is nil")
	}
	tag, err := m.pool.Exec(ctx, `DELETE FROM account_locks WHERE account = $1`, account)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrLockNotFound
	}
	return nil
}

// CleanupExpired 实现 LockManager.CleanupExpired。
// 返回被删除的行数，供定时任务上报指标用。
func (m *PGLockManager) CleanupExpired(ctx context.Context, now time.Time) (int, error) {
	if m.pool == nil {
		return 0, errors.New("login_security: pg pool is nil")
	}
	tag, err := m.pool.Exec(ctx, `DELETE FROM account_locks WHERE locked_until <= $1`, now)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// Compile-time guarantee.
var _ LockManager = (*PGLockManager)(nil)
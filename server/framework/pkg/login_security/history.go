package login_security

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HistoryRecorder 负责登录成功历史的写入与最近 N 次查询。
//
// 数据用于异地告警判定：取"该账号最近 N 次成功登录的 IP/device_id 列表"，
// 与本次登录的 IP/device_id 对比，决定是否触发 AnomalySignal。
type HistoryRecorder interface {
	// Record 写入一条成功登录记录。
	Record(ctx context.Context, entry LoginHistoryEntry) error

	// ListRecent 取账号最近 limit 条成功登录记录，按时间倒序。
	ListRecent(ctx context.Context, accountID uint, limit int) ([]LoginHistoryEntry, error)
}

// PGHistoryRecorder 是 HistoryRecorder 的 PostgreSQL 实现。
type PGHistoryRecorder struct {
	pool *pgxpool.Pool
}

// NewPGHistoryRecorder 构造基于 pgxpool 的 HistoryRecorder。
func NewPGHistoryRecorder(pool *pgxpool.Pool) *PGHistoryRecorder {
	return &PGHistoryRecorder{pool: pool}
}

// Record 实现 HistoryRecorder.Record。
func (r *PGHistoryRecorder) Record(ctx context.Context, e LoginHistoryEntry) error {
	if r.pool == nil {
		return errors.New("login_security: pg pool is nil")
	}
	loginAt := e.LoginAt
	if loginAt.IsZero() {
		loginAt = time.Now()
	}
	var userID any
	if e.UserID > 0 {
		userID = e.UserID
	}
	var tenantID any
	if e.TenantID > 0 {
		tenantID = e.TenantID
	}
	scope := string(e.Scope)
	if scope == "" {
		scope = string(ScopeTenant)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO login_history
			(account_id, user_id, tenant_id, scope, ip, user_agent, device_id, location, session_id, login_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), $10)
	`, e.AccountID, userID, tenantID, scope, e.IP, e.UserAgent, e.DeviceID, e.Location, e.SessionID, loginAt)
	return err
}

// ListRecent 实现 HistoryRecorder.ListRecent。
//
// 使用 idx_login_history_account_time 索引。
func (r *PGHistoryRecorder) ListRecent(ctx context.Context, accountID uint, limit int) ([]LoginHistoryEntry, error) {
	if r.pool == nil {
		return nil, errors.New("login_security: pg pool is nil")
	}
	if accountID == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT account_id, COALESCE(user_id, 0), COALESCE(tenant_id, 0), scope,
		       ip, COALESCE(user_agent, ''), COALESCE(device_id, ''),
		       COALESCE(location, ''), COALESCE(session_id, ''), login_at
		FROM login_history
		WHERE account_id = $1
		ORDER BY login_at DESC
		LIMIT $2
	`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LoginHistoryEntry
	for rows.Next() {
		var e LoginHistoryEntry
		var scope string
		if err := rows.Scan(&e.AccountID, &e.UserID, &e.TenantID, &scope,
			&e.IP, &e.UserAgent, &e.DeviceID, &e.Location, &e.SessionID, &e.LoginAt); err != nil {
			return nil, err
		}
		e.Scope = Scope(scope)
		out = append(out, e)
	}
	return out, rows.Err()
}

// Compile-time guarantee.
var _ HistoryRecorder = (*PGHistoryRecorder)(nil)
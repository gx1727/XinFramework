// Package session 提供会话生命周期管理（创建 / 校验 / 撤销）。
//
// 框架设计：
//   - 接口 SessionManager 定义能力，实现有 redisSessionManager（默认）与
//     dbSessionManager（Redis 不可用时 fallback）
//   - 进程级 defaultManager 由 Init 注入，业务代码通过 Manager() 拿
//   - Redis key 模板：sess:{SessionID} → JSON payload
//
// 会话与 JWT 的关系：
//   - JWT 中携带 SessionID，Auth 中间件在每次请求里调 Manager().Validate(SessionID)
//     检会话存活——这是“未撤销 + 未过期” 的唯一权威来源
//   - 退出登录调 Revoke(SessionID) 即可让该会话立即失效，无需等待 JWT 自然过期
package session

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/cache"
)

// sessionKeyPrefix 是 Redis 会话 key 的统一前缀。
const sessionKeyPrefix = "sess:"

// payload 是会话持久化到 Redis / DB 的载荷结构。
type payload struct {
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	TenantID  uint   `json:"tenant_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

// SessionManager 定义会话生命周期管理能力。
//
//   - Create：登录成功后写入会话
//   - Validate：Auth 中间件每个请求调用，存在且未过期返回 true
//   - Revoke：退出登录时调用，立刻让该 SessionID 失效
type SessionManager interface {
	Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
	Validate(sessionID string) (bool, error)
	Revoke(sessionID string) error
}

// defaultManager 是进程级会话管理器单例，由 Init 注入。
var defaultManager SessionManager

// Init 注入全局 SessionManager。boot.Init 阶段调用一次。
func Init(manager SessionManager) {
	defaultManager = manager
}

// Manager 返回当前 SessionManager。
func Manager() SessionManager {
	return defaultManager
}

type redisSessionManager struct{}

func (m *redisSessionManager) Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	if ttl <= 0 {
		return ErrInvalidSessionTTL
	}

	rdb := cache.Get()
	if rdb == nil {
		return ErrBackendUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p := payload{
		SessionID: sessionID,
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, sessionKeyPrefix+sessionID, b, ttl).Err()
}

func (m *redisSessionManager) Validate(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	rdb := cache.Get()
	if rdb == nil {
		return false, ErrBackendUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, err := rdb.Exists(ctx, sessionKeyPrefix+sessionID).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (m *redisSessionManager) Revoke(sessionID string) error {
	if sessionID == "" {
		return nil
	}

	rdb := cache.Get()
	if rdb == nil {
		return ErrBackendUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return rdb.Del(ctx, sessionKeyPrefix+sessionID).Err()
}

type dbSessionManager struct {
	pool      *pgxpool.Pool
	tableOnce sync.Once
}

func (m *dbSessionManager) ensureTable(ctx context.Context) {
	m.tableOnce.Do(func() {
		_, _ = m.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS auth_sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    role VARCHAR(64),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW())
`)
		_, _ = m.pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions (expires_at)`)
	})
}

func (m *dbSessionManager) Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	if ttl <= 0 {
		return ErrInvalidSessionTTL
	}

	ctx := context.Background()
	m.ensureTable(ctx)
	expiresAt := time.Now().Add(ttl)
	_, err := m.pool.Exec(ctx, `
		INSERT INTO auth_sessions (session_id, user_id, tenant_id, role, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id)
		DO UPDATE SET user_id = EXCLUDED.user_id, tenant_id = EXCLUDED.tenant_id, role = EXCLUDED.role, expires_at = EXCLUDED.expires_at
	`, sessionID, userID, tenantID, role, expiresAt)
	return err
}

func (m *dbSessionManager) Validate(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	ctx := context.Background()
	m.ensureTable(ctx)
	var cnt int64
	err := m.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM auth_sessions
		WHERE session_id = $1 AND expires_at > NOW()
	`, sessionID).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (m *dbSessionManager) Revoke(sessionID string) error {
	if sessionID == "" {
		return nil
	}

	ctx := context.Background()
	m.ensureTable(ctx)
	_, err := m.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE session_id = $1`, sessionID)
	return err
}

// NewRedisSessionManager creates a session manager using Redis
func NewRedisSessionManager() SessionManager {
	return &redisSessionManager{}
}

// NewDBSessionManager creates a session manager using database
func NewDBSessionManager(pool *pgxpool.Pool) SessionManager {
	return &dbSessionManager{pool: pool}
}

// Create delegates to defaultManager if available, otherwise returns error
func Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	if defaultManager != nil {
		return defaultManager.Create(sessionID, userID, tenantID, role, ttl)
	}
	return ErrBackendUnavailable
}

// Validate delegates to defaultManager if available, otherwise returns error
func Validate(sessionID string) (bool, error) {
	if defaultManager != nil {
		return defaultManager.Validate(sessionID)
	}
	return false, ErrBackendUnavailable
}

// Revoke delegates to defaultManager if available, otherwise returns error
func Revoke(sessionID string) error {
	if defaultManager != nil {
		return defaultManager.Revoke(sessionID)
	}
	return ErrBackendUnavailable
}

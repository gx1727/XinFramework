package session

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/cache"
)

const sessionKeyPrefix = "sess:"

type payload struct {
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	TenantID  uint   `json:"tenant_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

type SessionManager interface {
	Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
	Validate(sessionID string) (bool, error)
	Revoke(sessionID string) error
}

var (
	ensureTableOnce sync.Once
	defaultManager  SessionManager
)

func Init(manager SessionManager) {
	defaultManager = manager
}

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
	return cache.Get().Set(ctx, sessionKeyPrefix+sessionID, b, ttl).Err()
}

func (m *redisSessionManager) Validate(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, err := cache.Get().Exists(ctx, sessionKeyPrefix+sessionID).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (m *redisSessionManager) Revoke(sessionID string) error {
	if sessionID == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return cache.Get().Del(ctx, sessionKeyPrefix+sessionID).Err()
}

type dbSessionManager struct {
	pool *pgxpool.Pool
}

func (m *dbSessionManager) Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	if ttl <= 0 {
		return ErrInvalidSessionTTL
	}

	ctx := context.Background()
	ensureSessionTable(ctx, m.pool)
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
	ensureSessionTable(ctx, m.pool)
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
	ensureSessionTable(ctx, m.pool)
	_, err := m.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE session_id = $1`, sessionID)
	return err
}

func ensureSessionTable(ctx context.Context, pool *pgxpool.Pool) {
	ensureTableOnce.Do(func() {
		_, _ = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS auth_sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    role VARCHAR(64),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW())
`)
		_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions (expires_at)`)
	})
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

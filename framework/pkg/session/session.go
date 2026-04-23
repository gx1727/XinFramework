package session

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/db"
)

const sessionKeyPrefix = "sess:"

type payload struct {
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	TenantID  uint   `json:"tenant_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

var ensureTableOnce sync.Once

func Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	if sessionID == "" {
		return ErrEmptySessionID
	}
	if ttl <= 0 {
		return ErrInvalidSessionTTL
	}

	if rdb := cache.Get(); rdb != nil {
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

	pool := db.Get()
	if pool == nil {
		return ErrBackendUnavailable
	}
	ctx := context.Background()
	ensureSessionTable(ctx, pool)
	expiresAt := time.Now().Add(ttl)
	_, err := pool.Exec(ctx, `
		INSERT INTO auth_sessions (session_id, user_id, tenant_id, role, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id)
		DO UPDATE SET user_id = EXCLUDED.user_id, tenant_id = EXCLUDED.tenant_id, role = EXCLUDED.role, expires_at = EXCLUDED.expires_at
	`, sessionID, userID, tenantID, role, expiresAt)
	return err
}

func Validate(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}

	if rdb := cache.Get(); rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		n, err := rdb.Exists(ctx, sessionKeyPrefix+sessionID).Result()
		if err != nil {
			return false, err
		}
		return n == 1, nil
	}

	pool := db.Get()
	if pool == nil {
		return false, ErrBackendUnavailable
	}
	ctx := context.Background()
	ensureSessionTable(ctx, pool)
	var cnt int64
	err := pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM auth_sessions
		WHERE session_id = $1 AND expires_at > NOW()
	`, sessionID).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func Revoke(sessionID string) error {
	if sessionID == "" {
		return nil
	}

	if rdb := cache.Get(); rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return rdb.Del(ctx, sessionKeyPrefix+sessionID).Err()
	}

	pool := db.Get()
	if pool == nil {
		return nil
	}
	ctx := context.Background()
	ensureSessionTable(ctx, pool)
	_, err := pool.Exec(ctx, `DELETE FROM auth_sessions WHERE session_id = $1`, sessionID)
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
    created_at TIMESTAMPTZ DEFAULT NOW()
)`)
		_, _ = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions (expires_at)`)
	})
}

package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gx1727.com/xin/internal/infra/cache"
	"gx1727.com/xin/internal/infra/db"
	"gorm.io/gorm"
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
		return fmt.Errorf("empty session id")
	}
	if ttl <= 0 {
		return fmt.Errorf("invalid session ttl")
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

	d := db.Get()
	if d == nil {
		return fmt.Errorf("session backend unavailable: db not initialized")
	}
	ensureSessionTable(d)
	expiresAt := time.Now().Add(ttl)
	return d.Exec(`
INSERT INTO auth_sessions (session_id, user_id, tenant_id, role, expires_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (session_id)
DO UPDATE SET user_id = EXCLUDED.user_id, tenant_id = EXCLUDED.tenant_id, role = EXCLUDED.role, expires_at = EXCLUDED.expires_at
`, sessionID, userID, tenantID, role, expiresAt).Error
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

	d := db.Get()
	if d == nil {
		return false, fmt.Errorf("session backend unavailable: db not initialized")
	}
	ensureSessionTable(d)
	var cnt int64
	if err := d.Raw(`
SELECT COUNT(1)
FROM auth_sessions
WHERE session_id = ? AND expires_at > NOW()
`, sessionID).Scan(&cnt).Error; err != nil {
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

	d := db.Get()
	if d == nil {
		return nil
	}
	ensureSessionTable(d)
	return d.Exec(`DELETE FROM auth_sessions WHERE session_id = ?`, sessionID).Error
}

func ensureSessionTable(d *gorm.DB) {
	ensureTableOnce.Do(func() {
		_ = d.Exec(`
CREATE TABLE IF NOT EXISTS auth_sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL DEFAULT 0,
    role VARCHAR(64),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
)`).Error
		_ = d.Exec(`CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions (expires_at)`).Error
	})
}

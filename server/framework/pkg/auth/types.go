// Package auth exposes the public auth contract that business modules
// depend on. The concrete AccountRepository lives in apps/boot/auth —
// framework's user module (and any other consumer) only sees this
// interface, so the implementation can be swapped without touching
// downstream code.
//
// Phase 2 rationale: auth has moved from framework/internal/module/auth
// to apps/boot/auth. Business modules that need AccountRepository
// (notably user) cannot import apps/ directly because they live in
// the framework module. To bridge that gap, the interface definition
// stays public in framework/pkg/auth/, while the implementation lives
// in apps/boot/auth/.
package auth

import (
	"context"
	"time"
)

// Account is the global (cross-tenant) account record. Same struct
// shape as the one in apps/boot/auth — duplicated here to keep the
// framework module free of dependencies on the apps module.
type Account struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	RealName  string    `json:"real_name"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountAuth is a third-party authentication binding (wechat/qq/weibo).
type AccountAuth struct {
	ID         uint      `json:"id"`
	TenantID   uint      `json:"tenant_id"`
	AccountID  uint      `json:"account_id"`
	Type       string    `json:"type"`
	OpenID     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AccountRepository is the minimal subset of account data access that
// other business modules are allowed to depend on. The concrete
// implementation in apps/boot/auth satisfies this interface implicitly
// because the field types are identical.
type AccountRepository interface {
	GetByID(ctx context.Context, id uint) (*Account, error)
	GetByUsername(ctx context.Context, username string) (*Account, error)
	GetByPhone(ctx context.Context, phone string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*Account, error)
	Exists(ctx context.Context, account string) (bool, error)
}

// AccountAuthRepository is the third-party auth binding data access.
type AccountAuthRepository interface {
	GetByOpenID(ctx context.Context, tenantID uint, authType, openID string) (*AccountAuth, error)
	GetByAccountID(ctx context.Context, accountID uint) ([]AccountAuth, error)
	Create(ctx context.Context, tenantID, accountID uint, authType, openID, unionID, sessionKey string) (*AccountAuth, error)
	UpdateSessionKey(ctx context.Context, id uint, sessionKey string) error
	Delete(ctx context.Context, id uint) error
}
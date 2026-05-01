package auth

import (
	"context"
	"errors"
	"time"
)

// Account represents a global account (cross-tenant)
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

// AccountRepository defines data access operations for accounts
type AccountRepository interface {
	GetByID(ctx context.Context, id uint) (*Account, error)
	GetByUsername(ctx context.Context, username string) (*Account, error)
	GetByPhone(ctx context.Context, phone string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*Account, error)
	Exists(ctx context.Context, account string) (bool, error)
}

// AccountAuth represents a third-party authentication binding
type AccountAuth struct {
	ID         uint      `json:"id"`
	TenantID   uint      `json:"tenant_id"`
	AccountID  uint      `json:"account_id"`
	Type       string    `json:"type"` // wechat, qq, weibo, wxxcx
	OpenID     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AccountAuthRepository defines data access operations for account auths
type AccountAuthRepository interface {
	GetByOpenID(ctx context.Context, tenantID uint, authType, openID string) (*AccountAuth, error)
	GetByAccountID(ctx context.Context, accountID uint) ([]AccountAuth, error)
	Create(ctx context.Context, tenantID, accountID uint, authType, openID, unionID, sessionKey string) (*AccountAuth, error)
	UpdateSessionKey(ctx context.Context, id uint, sessionKey string) error
	Delete(ctx context.Context, id uint) error
}

var (
	ErrAccountNotFoundDB      = errors.New("account not found")
	ErrAccountAlreadyExistsDB = errors.New("account already exists")
	ErrAccountAuthNotFoundDB  = errors.New("account auth not found")
)

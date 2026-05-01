package user

import (
	"context"
	"errors"
	"time"
)

// User represents a user entity
type User struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	AccountID uint      `json:"account_id"`
	Code      string    `json:"code"`
	Nickname  string    `json:"nickname"`
	Status    int8      `json:"status"`
	RealName  string    `json:"real_name"`
	Avatar    string    `json:"avatar"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository defines data access operations for users
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error)
	Create(ctx context.Context, tenantID, accountID uint, code string) (*User, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	UpdatePhone(ctx context.Context, userID uint, phone string) error
	UpdateProfile(ctx context.Context, id uint, nickname, avatar string) error
	UpdateAvatar(ctx context.Context, id uint, avatar string) error
	Delete(ctx context.Context, id uint) error
}

var (
	ErrUserNotFoundDB = errors.New("user not found")
	ErrUserDisabledDB = errors.New("user is disabled")
)

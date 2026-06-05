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
	OrgID     *uint     `json:"org_id,omitempty"`
	OrgName   string    `json:"org_name,omitempty"`
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
	GetByIDScoped(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error)
	ListScoped(ctx context.Context, tenantID uint, keyword string, orgID *uint, page, size int) ([]User, int64, error)
	Create(ctx context.Context, tenantID, accountID uint, code string, orgID *uint) (*User, error)
	Update(ctx context.Context, id uint, req UpdateUserRepoReq) (*User, error)
	Patch(ctx context.Context, id uint, req PatchUserRepoReq) (*User, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	UpdateOrg(ctx context.Context, id uint, orgID *uint) (*User, error)
	UpdatePhone(ctx context.Context, userID uint, phone string) error
	UpdateProfile(ctx context.Context, id uint, nickname, avatar string) error
	UpdateAvatar(ctx context.Context, id uint, avatar string) error
	Delete(ctx context.Context, id uint) error
}

// UpdateUserRepoReq 全量更新请求
type UpdateUserRepoReq struct {
	Nickname string
	RealName string
	Avatar   string
	Status   int8
	OrgID    *uint
}

// PatchUserRepoReq 局部更新请求。nil 字段表示保持原值
type PatchUserRepoReq struct {
	Nickname *string
	RealName *string
	Avatar   *string
	Status   *int8
	OrgID    *uint
}

var (
	ErrUserNotFoundDB      = errors.New("user not found")
	ErrUserDisabledDB      = errors.New("user is disabled")
	ErrUserAlreadyExists   = errors.New("username already exists")
	ErrDefaultRoleNotFound = errors.New("default role not found")
)

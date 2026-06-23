package user

import (
	"context"
	"errors"

	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
)

// User struct 是跨 module 的 canonical 定义。
//
// 与任何外部消费者共享同一个 struct 类型，避免类型分裂。
type User = pkgrbac.User

// UserRepository 是 apps/tenant/user 的完整接口（包含 Scoped 变体、
// Create / Patch / UpdateProfile / UpdateAvatar / Delete 等仅本地
// 使用的方法）。PostgresUserRepository 实现此接口，同时**自动**
// 满足 pkgrbac.UserRepository（因为后者是前者的子集）。
//
// 不要把 UserRepository 别名到 pkgrbac.UserRepository——那会把本地
// 接口窄化，导致内部代码看不到 GetByIDScoped 等方法。
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByIDScoped(ctx context.Context, id uint) (*User, error)
	GetByAccount(ctx context.Context, tenantID, accountID uint) (*User, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*User, error)
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

// Compile-time guards live in repository.go where the concrete
// PostgresUserRepository is defined. The `NewUserRepository(db)` call
// returns a value assigned to UserRepository — if PostgresUserRepository
// fails to satisfy UserRepository OR pkgrbac.UserRepository, the
// registration call in module.go will fail to compile.

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
package user

import (
	"context"
	"errors"

	pkgrbac "gx1727.com/xin/framework/pkg/tenant/auth"
)

// User is a type alias of pkgrbac.User (= auth.User after 0023.3).
//
// apps/tenant/user does NOT redefine the User struct; it aliases
// the framework contract to keep consumers (cms/handler, weixin/service,
// extapi.User, etc.) working without type changes.
//
// Struct composition (locked in framework/pkg/tenant/auth/user.go):
//
//	identity.User (10 fields)  - cross-domain base
//	+ TenantID                   - tenant-domain unique
//	+ Phone, Email, OrgName      - tenant-domain JOIN fields
//	+ MarshalJSON()              - byte-level legacy JSON output
//
// apps/tenant/user contributes:
//
//	- UserRepository interface (full CRUD + Scoped variants)
//	- Update/Patch request DTOs
//	- DB error sentinels
//
// The local UserRepository is a superset of framework/pkg/tenant/auth.UserRepository;
// PostgresUserRepository satisfies both by Go interface satisfaction rules.
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
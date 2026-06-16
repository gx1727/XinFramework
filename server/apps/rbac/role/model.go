package role

import (
	"context"
	"errors"

	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
)

// 类型别名 —— Role struct 与 pkgrbac.Role 共享
type Role = pkgrbac.Role

// RoleRepository 是 apps/rbac/role 的完整接口（包含 Create / Update /
// Patch / Delete 等本地方法）。PostgresRoleRepository 同时满足本接口和
// pkgrbac.RoleRepository（后者是前者的子集）。
//
// 不要 alias 到 pkgrbac.RoleRepository——会窄化本地接口。
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Role, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Role, int64, error)
	Create(ctx context.Context, tenantID uint, req CreateRoleRepoReq) (*Role, error)
	Update(ctx context.Context, id uint, req UpdateRoleRepoReq) (*Role, error)
	Patch(ctx context.Context, id uint, req PatchRoleRepoReq) (*Role, error)
	Delete(ctx context.Context, id uint) error
}

// Compile-time guards live in repository.go where PostgresRoleRepository
// is defined. module.go's pkgrbac.RegisterRoleRepository call will fail
// to compile if PostgresRoleRepository doesn't satisfy both interfaces.

// CreateRoleRepoReq fields for role creation
type CreateRoleRepoReq struct {
	Code        string
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
}

// UpdateRoleRepoReq fields for role update
type UpdateRoleRepoReq struct {
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
}

// PatchRoleRepoReq 局部更新请求。nil 字段表示保持原值不更新
type PatchRoleRepoReq struct {
	Name        *string
	Description *string
	DataScope   *int8
	IsDefault   *bool
	Sort        *int
	Status      *int8
}

var (
	ErrRoleNotFoundDB        = errors.New("role not found")
	ErrDefaultRoleNotFoundDB = errors.New("default role not found")
)
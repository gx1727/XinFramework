// Package sysuser 实现"平台域用户身份"管理 API（sys_users 表）。
//
// 与 apps/platform/sys_menu 同属 platform 域。关键不变量：
//
//	| 维度          | apps/tenant/user (rbac)       | apps/platform/sys_user     |
//	| tenant_id     | 来自 ctx.TenantID             | 不存在（平台域无 tenant）   |
//	| DB 事务上下文 | RunInTenantTx                 | RunInPlatformTx (bypass)   |
//	| 中间件        | Require(permission.P(...))    | RequirePlatformRole("super_admin") |
//	| 路由前缀      | /api/v1/users                 | /api/v1/platform/sys-users |
//	| 错误码段      | 5001-5099                     | 15100-15199                |
//
// 抽象层级：
//
//	  framework/pkg/identity.User            // 跨域基类
//	            ↑ embedded by
//	  framework/pkg/platformauth.User        // 平台域 User
//	            ↑ used by
//	  apps/platform/sys_user.User            // 本模块内部 Go struct
//
// 字段与 tenant_users 完全对齐（一个 account_id 对应一个 sys_user，
// 不带 tenant_id；一个 account 也可对应 0..N 个 tenant_user）。
package sysuser

import (
	"context"
	"time"
)

// User 是本模块的内部 Go struct，包装 platformauth.User。
// 字段集与 platformauth.User 完全相同（identity.User 子集）。
// 保持本地 struct 的原因：未来加平台专属字段时（platform_level、scope_tags
// 等）只影响本包，不污染 framework contracts。
type User struct {
	ID        uint       `json:"id"`
	AccountID uint       `json:"account_id"`
	OrgID     *uint      `json:"org_id"`
	Code      string     `json:"code"`
	RealName  string     `json:"real_name"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Status    int8       `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Roles     []RoleLite `json:"roles,omitempty"` // 关联 sys_roles 摘要
}

// RoleLite 是关联 sys_roles 的最小投影（避免在用户列表 N+1）。
type RoleLite struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// CreateRepoReq 是 repository 层的入参（不含任何 API 层字段）。
type CreateRepoReq struct {
	AccountID uint
	Code      string
	OrgID     *uint
	RealName  string
	Nickname  string
	Avatar    string
	Status    int8
	CreatedBy uint
}

// UpdateRepoReq 字段语义：零值表示"不更新"（除 ID 必填外）。
// Caller 负责把"不改的字段"留零。
type UpdateRepoReq struct {
	Code      *string
	OrgID     *uint
	RealName  *string
	Nickname  *string
	Avatar    *string
	Status    *int8
	UpdatedBy uint
}

// Repository 是 sys_user 的数据访问契约。
//
// 所有方法必须在 db.RunInPlatformTx 上下文中调用（不依赖 caller，
// 但 platformauth contract 不强制 —— 是约定的"不变量"）。
// 与 apps/tenant/user.UserRepository 形状一致，无 tenantID 参数。
type Repository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, keyword string, page, size int) ([]User, int64, error)
	Create(ctx context.Context, req CreateRepoReq) (*User, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*User, error)
	UpdateStatus(ctx context.Context, id uint, status int8, updatedBy uint) error
	Delete(ctx context.Context, id uint, updatedBy uint) error
	ListRoles(ctx context.Context, userID uint) ([]RoleLite, error)
	GrantRole(ctx context.Context, userID, roleID uint) error
	RevokeRole(ctx context.Context, userID, roleID uint) error
}

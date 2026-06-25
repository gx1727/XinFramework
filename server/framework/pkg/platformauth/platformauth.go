// Package platformauth 暴露平台域公开契约，供平台模块（apps/platform/sys_*）
// 与仅限平台的框架辅助使用，用于操作平台身份套件
// （sys_user / sys_role / sys_menu / sys_permission / sys_org）。
//
// 具体实现在 apps/platform/sys_<name>/。apps/platform/ 之外的模块只能
// 依赖本包，不能 import apps/。跨域消费者（weixin 模块、未来的仅平台模块）
// 都走这些接口。
//
// 平台域规则（Phase 0023）：
//   - 平台域无 tenant_id。sys_users / sys_roles / sys_menus / sys_permissions /
//     sys_orgs 都是单租户的。
//   - 平台域不启用 RLS。鉴权通过 API 层的 RequirePlatformRole(super_admin) +
//     db.RunInPlatformTx 上下文标记来强制。
//   - 一个 accounts 行可以持有 0 或 1 个 sys_users 行（每个全局账号最多一个
//     平台身份）。
package platformauth

import (
	"context"

	"gx1727.com/xin/framework/pkg/identity"
)

// User 是平台域用户。通过嵌入 identity.User 让公共字段保持单点；
// 平台侧目前不增加平台特有字段。如需增加（如 platform_level），
// 直接在本结构添加即可。
type User struct {
	identity.User
}

// Role 是平台域角色，嵌入 identity.Role。
//
// 平台侧 DataScope 语义：
//   1 = ALL              — 可看所有 sys_* 行
//   2 = SELF             — 仅看自己创建的行
//   4 = ORG_AND_CHILDREN — 看自己部门 + 子部门（默认未设置时）
type Role struct {
	identity.Role
	// Extend 故意不放进 identity.Role，避免基础结构被 JSONB 形字段污染。
	// 平台角色如果未来需要 extend，统一放在本包装结构里。
	Extend map[string]any `json:"extend,omitempty"`
}

// Menu 是平台域菜单，嵌入 identity.Menu。
type Menu struct {
	identity.Menu
}

// Permission 是平台域权限，嵌入 identity.Permission。
type Permission struct {
	identity.Permission
}

// Org 是平台域组织，嵌入 identity.Org。
type Org struct {
	identity.Org
}

// UserRepository 是跨模块的平台用户契约。
//
// 方法签名镜像租户侧 UserRepository 但去掉 tenantID 参数——平台域无租户。
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, keyword string, page, size int) ([]User, int64, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
}

// RoleRepository 是跨模块的平台角色契约。
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, code string) (*Role, error)
	List(ctx context.Context, keyword string, page, size int) ([]Role, int64, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	Grant(ctx context.Context, userID, roleID uint) error
	Revoke(ctx context.Context, userID, roleID uint) error
}

// MenuRepository 是跨模块的平台菜单契约。
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, code string) (*Menu, error)
	List(ctx context.Context, keyword string, page, size int) ([]Menu, int64, error)
	Tree(ctx context.Context) ([]Menu, error)
}

// PermissionRepository 是跨模块的平台权限契约。
type PermissionRepository interface {
	GetByID(ctx context.Context, id uint) (*Permission, error)
	GetByCode(ctx context.Context, code string) ([]Permission, error)
	List(ctx context.Context, menuID *uint, keyword string, page, size int) ([]Permission, int64, error)
}

// OrgRepository 是跨模块的平台组织契约。
// 当前只有 GetByID，具体方法会在 Phase 0023.1 业务需求确认后补充。
type OrgRepository interface {
	GetByID(ctx context.Context, id uint) (*Org, error)
}

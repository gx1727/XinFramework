// Package platformauth exposes the public contracts that platform
// modules (apps/platform/sys_*) and platform-only framework helpers
// use to interact with the platform identity suite
// (sys_user / sys_role / sys_menu / sys_permission / sys_org).
//
// The concrete implementations live in apps/platform/sys_<name>/. Apps
// outside of apps/platform/ must depend only on this pkg, not on
// apps/. Cross-domain consumers (the weixin module, future
// platform-only consumers) go through these interfaces.
//
// Domain rule (Phase 0023):
//   - The platform domain has no tenant_id. sys_users, sys_roles,
//     sys_menus, sys_permissions, sys_orgs are all single-tenant.
//   - The platform domain does NOT enable RLS. Security is enforced
//     at the API layer by RequirePlatformRole(super_admin) + the
//     db.RunInPlatformTx context marker.
//   - One accounts row can hold one sys_users row (one platform
//     identity per global login).
package platformauth

import (
	"context"
	"time"

	"gx1727.com/xin/framework/pkg/identity"
)

// User is the platform-domain user. It embeds identity.User so the
// common fields stay in one place; the platform side adds no
// platform-only fields today. When a platform-only field is needed
// (e.g. platform_level), add it here directly.
type User struct {
	identity.User
}

// Role is the platform-domain role. Embeds identity.Role.
//
// DataScope semantics on the platform side:
//   1 = ALL              — see every sys_* row
//   2 = SELF             — see only own created rows
//   4 = ORG_AND_CHILDREN — see own org + sub-orgs (default if not set)
type Role struct {
	identity.Role
	// Extend is intentionally not part of identity.Role to keep the
	// base struct free of JSONB-shaped fields. Platform roles keep
	// extend in this wrapper if/when it is needed.
	Extend map[string]any `json:"extend,omitempty"`
}

// Menu is the platform-domain menu. Embeds identity.Menu.
type Menu struct {
	identity.Menu
}

// Permission is the platform-domain permission. Embeds
// identity.Permission.
type Permission struct {
	identity.Permission
}

// Org is the platform-domain org. Embeds identity.Org.
type Org struct {
	identity.Org
}

// UserRepository is the cross-module platform user contract.
//
// Methods mirror the tenant-side UserRepository shape with the
// tenantID parameter removed — the platform domain has no tenant.
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, keyword string, page, size int) ([]User, int64, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
}

// RoleRepository is the cross-module platform role contract.
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, code string) (*Role, error)
	List(ctx context.Context, keyword string, page, size int) ([]Role, int64, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	Grant(ctx context.Context, userID, roleID uint) error
	Revoke(ctx context.Context, userID, roleID uint) error
}

// MenuRepository is the cross-module platform menu contract.
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, code string) (*Menu, error)
	List(ctx context.Context, keyword string, page, size int) ([]Menu, int64, error)
	Tree(ctx context.Context) ([]Menu, error)
}

// PermissionRepository is the cross-module platform permission contract.
type PermissionRepository interface {
	GetByID(ctx context.Context, id uint) (*Permission, error)
	GetByCode(ctx context.Context, code string) ([]Permission, error)
	List(ctx context.Context, menuID *uint, keyword string, page, size int) ([]Permission, int64, error)
}

// OrgRepository is the cross-module platform org contract.
// Returns an empty interface today — concrete methods will be added
// in Phase 0023.1 once business requirements are confirmed.
type OrgRepository interface {
	GetByID(ctx context.Context, id uint) (*Org, error)
}

// Time aliases for callers that prefer the platform-domain time
// constants. Kept here so the platform package can stay self-contained
// without leaking identity internals.
type (
	_ = time.Time
)

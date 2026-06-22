// Package identity holds the field-level base types shared by the
// platform domain (sys_*) and the tenant domain (tenant_*). The
// platform and tenant packages each define their own Go struct
// (platformauth.User, tenant/auth.User, …) that embeds the
// corresponding identity struct, so cross-domain consumers can read
// the common fields without depending on either side.
//
// Why a base package?
//
//   - One source of truth for "what is a User / Role / Menu /
//     Permission" across domains. When a field is added, the
//     identity struct is updated first, then both domains pick it up
//     via embedding.
//   - Stable contract: identity does not import framework/internal
//     or apps/, so it sits in pkg/ and can be consumed by both layers.
//
// When NOT to add a field here:
//
//   - If the field is platform-only (e.g. "platform_level") put it on
//     platformauth.User directly.
//   - If the field is tenant-only (e.g. TenantID) put it on the
//     tenant-side struct directly.
package identity

import "time"

// User is the cross-domain base for a user identity entity.
//
// Platform users (sys_users) and tenant users (tenant_users) both
// expose these fields. The only true difference between the two
// domains is that the tenant side has an extra TenantID; the
// platform side does not.
type User struct {
	ID        uint      `json:"id"`
	AccountID uint      `json:"account_id"`
	OrgID     *uint     `json:"org_id"`
	Code      string    `json:"code"`
	RealName  string    `json:"real_name"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Org is the cross-domain base for an organization / department.
type Org struct {
	ID          uint      `json:"id"`
	ParentID    *uint     `json:"parent_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AdminCode   string    `json:"admin_code"`
	Ancestors   string    `json:"ancestors"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role is the cross-domain base for a role.
type Role struct {
	ID          uint      `json:"id"`
	OrgID       *uint     `json:"org_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DataScope   int8      `json:"data_scope"`
	IsDefault   bool      `json:"is_default"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Menu is the cross-domain base for a navigation menu.
//
// Both domains expose the same fields. Platform menus live in
// sys_menus; tenant menus in tenant_menus.
type Menu struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Subtitle  string    `json:"subtitle"`
	URL       string    `json:"url"`
	Path      string    `json:"path"`
	Icon      string    `json:"icon"`
	Sort      int       `json:"sort"`
	ParentID  *uint     `json:"parent_id"`
	Ancestors string    `json:"ancestors"`
	Visible   bool      `json:"visible"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Permission is the cross-domain base for an action permission
// (e.g. "user:list", "tenant:create").
//
// Platform permissions live in sys_permissions; tenant permissions
// live in tenant_permissions (renamed from resources). The MenuID
// field is the owning menu, mirroring the resources.menu_id FK.
type Permission struct {
	ID          uint      `json:"id"`
	MenuID      *uint     `json:"menu_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

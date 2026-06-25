// Package identity 定义平台域（sys_*）与租户域（tenant_*）共享的字段级基础类型。
// 平台与租户包各自定义自己的 Go 结构（platformauth.User / tenant/auth.User 等），
// 通过嵌入对应的 identity 结构，让跨域消费者可以读取公共字段而不依赖任一侧。
//
// 为什么要有 identity 这个基础包？
//
//   - “用户 / 角色 / 菜单 / 权限是什么”的唯一事实来源。新增字段时，
//     先改 identity，两侧通过嵌入自动跟随。
//   - 稳定契约：identity 不依赖 framework/internal 或 apps/，位于 pkg/
//     可以被两层同时消费。
//
// 哪些字段不应加在这里：
//   - 仅平台域（如 platform_level）请直接放在 platformauth.User
//   - 仅租户域（如 TenantID）请直接放在租户侧的结构
package identity

import "time"

// User 是跨域的用户身份基础类型。
//
// 平台用户（sys_users）与租户用户（tenant_users）都暴露这些字段。
// 两域的唯一真实差异是租户侧多一个 TenantID，平台侧没有。
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

// Org 是跨域的组织 / 部门基础类型。
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
// live in tenant_permissions (renamed from tenant_permissions). The MenuID
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

package auth

import "gx1727.com/xin/framework/pkg/resp"

// LoginScope 登录作用域：tenant（业务租户登录） / platform（平台管理员登录）
//
// 决定 token 中 tenant_id 含义与后续能访问的路由空间：
//   - tenant   → tenant_id > 0，可访问 /api/v1/t/*（业务域）
//   - platform → tenant_id = 0，可访问 /api/v1/admin/*（平台域）
type LoginScope string

const (
	LoginScopeTenant   LoginScope = "tenant"
	LoginScopePlatform LoginScope = "platform"
)

// 平台登录专用错误（独立错误码段，不和 tenant 登录冲突）
var (
	ErrPlatformLoginAccountNotFound = resp.Err(1012, "平台账号不存在")
	ErrPlatformLoginNotAdmin        = resp.Err(1013, "该账号不具备平台管理员权限")
	ErrPlatformLoginDisabled        = resp.Err(1014, "平台账号已禁用")
)

// tenantLoginRequest 租户域登录（业务用户）
type tenantLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID uint   `json:"tenant_id" binding:"required"`
}

// platformLoginRequest 平台域登录（super_admin 等）
//
// 不带 tenant_id：平台管理员不属于任何租户；登录后整个会话不出现 tenant 概念。
// 后端验证账号 + 密码 + 至少有一个 platform_role（如 super_admin）。
type platformLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// User 是登录/注册响应里的精简用户视图，对应前端的 NavUser 字段。
//
// 注：保留 Code/Role 是因为前端 authStore.User 已依赖；Nickname/RealName/Avatar/Email
// 用于侧边栏展示。RealName 优先于 Nickname 作为显示名（前端会自己 fallback）。
//
// 字段含义按 Scope 略有差异：
//   - scope=tenant   → TenantID > 0，Role 是租户角色 code（如 "admin"）
//   - scope=platform → TenantID = 0，Role 为空（平台管理员无租户角色），PlatformRoles 必有值
type User struct {
	ID       uint   `json:"id"`
	TenantID uint   `json:"tenant_id"`
	Code     string `json:"code"`
	Role     string `json:"role"`

	Nickname string `json:"nickname,omitempty"`
	RealName string `json:"real_name,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Email    string `json:"email,omitempty"`

	// PlatformRoles 平台级角色列表（super_admin 等），与 JWT claims 对齐，
	// 让前端能在不解析 JWT 的前提下判断是否能访问 /admin/* 平台域路由。
	PlatformRoles []string `json:"platform_roles,omitempty"`
}

type LoginResult struct {
	Token        string
	RefreshToken string
	Scope        LoginScope
	User         User
}

type registerRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	TenantID uint   `json:"tenant_id" binding:"required"`
	RealName string `json:"real_name"`
}

type registerResult struct {
	Token        string
	RefreshToken string
	Scope        LoginScope
	User         User
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshResult struct {
	Token        string
	RefreshToken string
}
package auth

import (
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/resp"
)

// LoginScope 登录作用域：tenant（业务租户登录） / platform（平台管理员登录）
//
// 决定 token 中 tenant_id 含义与后续能访问的路由空间：
//   - tenant   → tenant_id > 0，可访问 /api/v1/* 业务域（Auth + RequireTenantContext）
//   - platform → tenant_id = 0，可访问 /api/v1/platform/* 平台域（Auth + RequirePlatformRole）
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

// 登录前置检查专用错误
var (
	ErrNoLoginIdentity = resp.Err(1015, "账号无可用登录身份（无 tenant 身份且无 platform 角色）")
)

// Refresh 切租户专用错误
var (
	ErrCrossTenantSwitchFromPlatform = resp.Err(1016, "平台 token 不能切换到租户，请用 platform-login 重新登录")
)

// tenantLoginRequest 租户域登录（业务用户）
type tenantLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID uint   `json:"tenant_id" binding:"required"`
}

// loginPrecheckRequest 登录前置检查（账号密码 → 列出可用身份）。
//
// 用于"多身份账号"登录流程：账号可能在多个租户都有 users 记录。
// 前端提交账号密码，服务器验证后返回所有可用身份，前端让用户选择后再
// 调用 /auth/select-tenant（等价于 /auth/tenant-login）签发 token。
//
// 单身份账号可以跳过此步直接调 /auth/tenant-login 或 /auth/platform-login。
type loginPrecheckRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// loginPrecheckResult 登录前置检查的响应。
//
// 设计要点：
//   - 不签发 token：前端调 precheck 后还要调 select-tenant 才拿到 token
//   - platform_available: true 时前端可选择调 /auth/platform-login（不走 tenant）
//   - tenant_identities 为空 + platform_available 为 false：账号无登录权限（403）
//   - tenant_identities 为空 + platform_available 为 true：纯平台账号，precheck 后调 platform-login
//   - tenant_identities 非空：列出所有可选身份，每个都能用 /auth/select-tenant 登录
//
// security note：响应里不含密码或敏感字段，仅展示用。
type loginPrecheckResult struct {
	AccountID         uint                     `json:"account_id"`
	AccountStatus     int8                     `json:"account_status"`
	RealName          string                   `json:"real_name,omitempty"`
	Email             string                   `json:"email,omitempty"`
	PlatformAvailable bool                     `json:"platform_available"`
	PlatformRoles     []string                 `json:"platform_roles,omitempty"`
	TenantIdentities  []pkgauth.TenantIdentity `json:"tenant_identities"`
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

	// Permissions 当前用户持有的资源权限码（"resource:action" 形式）。
	// 0024+：前端用此字段做按钮可见性 / 路由守门，避免每个页面 round-trip
	// 到 /permissions/me。权威校验仍在后端中间件（Require(P(Res, Act))）。
	// 为 nil 时视为"零权限"，前端用空数组处理即可。
	Permissions []string `json:"permissions,omitempty"`
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
	// TenantID 可选；用于 Refresh 切租户（路径 B 多身份支持）。
	//   - 0 或缺省：刷当前租户，沿用 refresh token 里的 TenantID（向后兼容）
	//   - 非 0：切到新租户，账号必须在新租户有 users 记录，否则 403
	//   - 不允许从 platform token 切到租户（platform token 没有 users 上下文）
	TenantID uint `json:"tenant_id,omitempty"`
}

type refreshResult struct {
	Token        string
	RefreshToken string
}

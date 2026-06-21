package auth

type loginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID uint   `json:"tenant_id" binding:"required"`
}

// User 是登录/注册响应里的精简用户视图，对应前端的 NavUser 字段。
//
// 注：保留 Code/Role 是因为前端 authStore.User 已依赖；Nickname/RealName/Avatar/Email
// 用于侧边栏展示。RealName 优先于 Nickname 作为显示名（前端会自己 fallback）。
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
	User         User
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshResult struct {
	Token        string
	RefreshToken string
}

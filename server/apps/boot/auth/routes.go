package auth

import (
	"github.com/gin-gonic/gin"
)

// Register 把 auth 模块路由挂到 public/protected 两个 RouterGroup 上。
//
// 登录入口按 scope 拆开：
//   - POST /auth/tenant-login      租户域登录（业务用户，需要 tenant_id）
//   - POST /auth/platform-login    平台域登录（super_admin 等，不传 tenant）
//   - POST /auth/login-precheck    登录前置检查（多身份账号，列可选身份）
//   - POST /auth/select-tenant     选择 tenant 身份签 token（等价 tenant-login）
//
// 保留旧 /auth/login 兼容期（路由层转发 → /auth/tenant-login），老调用方短期不挂。
func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	// 主登录入口
	public.POST("/auth/tenant-login", h.TenantLogin)
	public.POST("/auth/platform-login", h.PlatformLogin)

	// 路径 B 多身份支持
	public.POST("/auth/login-precheck", h.LoginPrecheck)
	public.POST("/auth/select-tenant", h.SelectTenant)

	public.POST("/auth/register", h.Register)
	public.POST("/auth/refresh", h.Refresh)

	// 兼容期：旧 /auth/login 重定向到新入口
	public.POST("/auth/login", h.legacyLoginRedirect)

	protected.POST("/auth/logout", h.Logout)
}

// legacyLoginRedirect 兼容期过渡：旧 /auth/login 自动跳到 /auth/tenant-login。
// 给老 SDK / curl 测试一段时间缓冲；移除时机由后续 phase 决定。
func (h *Handler) legacyLoginRedirect(c *gin.Context) {
	// 简化实现：直接转发到 TenantLogin（行为等价于 /auth/tenant-login）
	h.TenantLogin(c)
}
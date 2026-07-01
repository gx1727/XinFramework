package auth

import (
	"github.com/gin-gonic/gin"
)

// Register 把 auth 模块路由挂到 public/protected 两个 RouterGroup 上。
//
// 登录入口按 scope 拆开：
//   - POST /auth/tenant-login      租户域登录（业务用户，需要 tenant_id）
//   - POST /auth/sys-login         sys 域登录（super_admin 等，不传 tenant）
//   - POST /auth/login-precheck    登录前置检查（多身份账号，列可选身份）
//   - POST /auth/select-tenant     选择 tenant 身份签 token（等价 tenant-login）
func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	// 主登录入口
	public.POST("/auth/tenant-login", h.TenantLogin)
	public.POST("/auth/sys-login", h.SysLogin)

	// 路径 B 多身份支持
	public.POST("/auth/login-precheck", h.LoginPrecheck)
	public.POST("/auth/select-tenant", h.SelectTenant)

	public.POST("/auth/register", h.Register)
	public.POST("/auth/refresh", h.Refresh)

	protected.POST("/auth/logout", h.Logout)
}

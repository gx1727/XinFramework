package tenants

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// PlatformRoleSuperAdmin 平台域"可以模拟登录租户后台"的专属角色名。
//
// 0024+ 终态：这是 super_admin 唯一保留的特殊用法——决定 platform user 能否
// 进 tenant 后台。其他平台域操作（租户 CRUD / 菜单 / 用户 / 角色）走纯 RBAC，
// 不再因为持有本角色而获得隐式全权限。
const (
	PlatformRoleSuperAdmin = "super_admin"
)

// Register 把平台租户管理路由挂到 /api/v1/platform/tenants。
//
// 0024+ 中间件链：
//  1. protected.Use(middleware.Auth(...))              来自 framework.go：注入 XinContext
//  2. g := protected.Group("/tenants",
//     pkgmiddleware.RequireAnyPlatformRole())            挡住非 platform 用户
//  3. 各路由上叠加 Require(P(ResTenant, ...)) 做资源级权限细分
//
// Impersonate 是 super_admin 唯一专属的端点，独立分组 + RequirePlatformRole(super_admin)
// 守卫（这是 super_admin 唯一保留的"按角色硬编码"用法）。
func Register(protected *gin.RouterGroup, h *Handler) {
	// 租户 CRUD：任何 platform 角色都可调到这里；具体能力由 ResTenant:* 资源权限码决定
	g := protected.Group("/tenants",
		pkgmiddleware.RequireAnyPlatformRole(),
	)
	{
		g.POST("", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActCreate)), h.Create)
		g.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActUpdate)), h.Update)
		g.PUT("/:id/status", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActUpdate)), h.UpdateStatus)
		g.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActDelete)), h.Delete)
		// Purge 单独 endpoint：硬删是不可逆操作，URL 用动词区分，避免误用 DELETE 触发
		g.POST("/:id/purge", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActDelete)), h.Purge)
		g.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.Get)
		g.GET("", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.List)

		// Impersonate 模拟登录：super_admin 专属（决定能否进租户后台）。
		// 0024+：这是 super_admin 唯一保留的"按角色硬编码"用法。
		// 在 g 下面单 route 叠加 RequirePlatformRole(super_admin)，与 group 级
		// RequireAnyPlatformRole 形成"platform + super_admin"双门卫。
		g.POST("/:id/impersonate",
			pkgmiddleware.RequirePlatformRole(PlatformRoleSuperAdmin),
			h.Impersonate,
		)
	}
}

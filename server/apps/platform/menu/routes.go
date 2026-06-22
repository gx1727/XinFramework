package menu

import (
	"github.com/gin-gonic/gin"

	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

// Register 把平台菜单路由挂到 /api/v1/platform/menus。
//
// 中间件约束（顺序自上而下）：
//
//	1. protected.Use(middleware.Auth(...))            // 来自 framework.go
//	   ↓
//	2. g := protected.Group("/menus",
//	                    pkgmiddleware.RequirePlatformRole("super_admin"))
//	   ↓
//	3. 各路由上不再重复挂 RequirePlatformRole —— group 级已守卫
//
// 如果调用方忘了 super_admin 身份：
//   - middleware.Auth 先注入 XinContext
//   - RequirePlatformRole 在 Auth 之后执行，检查 PlatformRoles
//   - 没 super_admin 直接 403 "需要平台级角色"
//
// **没有额外前缀**：`/api/v1` 版本化空间内，`/platform` 子空间表示
// "平台管理域"。这与 `/api/v1/menus`（租户域）形成清晰边界。
func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/menus",
		pkgmiddleware.RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin),
	)
	{
		g.GET("", h.List)
		g.GET("/tree", h.Tree)
		g.GET("/:id", h.Get)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
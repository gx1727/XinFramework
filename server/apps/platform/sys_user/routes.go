package sysuser

import (
	"github.com/gin-gonic/gin"

	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

// Register 把 sys_user 路由挂到 /api/v1/platform/sys-users。
//
// 中间件链：
//   - protected.Use(middleware.Auth(...))      来自 framework.go（已挂）
//   - group 级 RequirePlatformRole(super_admin)
//   - 各 route 不再重复挂
func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/sys-users",
		pkgmiddleware.RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin),
	)
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.PUT("/:id/status", h.UpdateStatus)
		g.DELETE("/:id", h.Delete)
		g.PUT("/:id/roles", h.AssignRoles)
	}
}

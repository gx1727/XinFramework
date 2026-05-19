package permission

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	// Permission management for roles
	protected.GET("/roles/:id/permissions", middleware.RequirePermission("role", "list"), h.GetPermissions)
	protected.POST("/roles/:id/permissions", middleware.RequirePermission("role", "update"), h.AssignPermissions)
	protected.PUT("/roles/:id/permissions", middleware.RequirePermission("role", "update"), h.AssignPermissions)
	protected.GET("/roles/:id/menus", middleware.RequirePermission("role", "list"), h.GetMenus)
	protected.GET("/roles/:id/resources", middleware.RequirePermission("role", "list"), h.GetResources)
}

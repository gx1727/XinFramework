package permission

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	// Permission management for roles (resources only - menus moved to role module)
	protected.GET("/roles/:id/permissions", middleware.RequirePermission(permission.ResRole, permission.ActList), h.GetPermissions)
	protected.POST("/roles/:id/permissions", middleware.RequirePermission(permission.ResRole, permission.ActUpdate), h.AssignPermissions)
	protected.PUT("/roles/:id/permissions", middleware.RequirePermission(permission.ResRole, permission.ActUpdate), h.AssignPermissions)
	protected.GET("/roles/:id/resources", middleware.RequirePermission(permission.ResRole, permission.ActList), h.GetResources)
}

package permission

import (
	"github.com/gin-gonic/gin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	// Permission management for roles
	protected.GET("/roles/:id/permissions", h.GetPermissions)
	protected.POST("/roles/:id/permissions", h.AssignPermissions)
	protected.PUT("/roles/:id/permissions", h.AssignPermissions)
	protected.GET("/roles/:id/menus", h.GetMenus)
	protected.GET("/roles/:id/resources", h.GetResources)
}

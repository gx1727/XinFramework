package permission

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	// Permission management for roles
	protected.GET("/roles/:id/permissions", h.GetPermissions)
	protected.POST("/roles/:id/permissions", h.AssignPermissions)
	protected.PUT("/roles/:id/permissions", h.AssignPermissions)
	protected.GET("/roles/:id/menus", h.GetMenus)
	protected.GET("/roles/:id/resources", h.GetResources)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("permission", func(_, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

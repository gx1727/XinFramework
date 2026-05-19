package role

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	roles := protected.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission("role", "list"), h.List)
		roles.GET("/:id", middleware.RequirePermission("role", "list"), h.Get)
		roles.POST("", middleware.RequirePermission("role", "create"), h.Create)
		roles.PUT("/:id", middleware.RequirePermission("role", "update"), h.Update)
		roles.DELETE("/:id", middleware.RequirePermission("role", "delete"), h.Delete)
		roles.GET("/:id/data-scopes", middleware.RequirePermission("role", "list"), h.GetDataScopes)
		roles.PUT("/:id/data-scopes", middleware.RequirePermission("role", "update"), h.UpdateDataScopes)
	}
}

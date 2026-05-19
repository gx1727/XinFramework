package resource

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	resources := protected.Group("/resources")
	{
		resources.GET("", middleware.RequirePermission("resource", "list"), h.List)
		resources.GET("/:id", middleware.RequirePermission("resource", "list"), h.Get)
		resources.POST("", middleware.RequirePermission("resource", "create"), h.Create)
		resources.PUT("/:id", middleware.RequirePermission("resource", "update"), h.Update)
		resources.DELETE("/:id", middleware.RequirePermission("resource", "delete"), h.Delete)
		resources.GET("/by-menu/:menu_id", middleware.RequirePermission("resource", "list"), h.GetByMenu)
	}
}

package tenant

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	tenants := protected.Group("/tenants")
	{
		tenants.POST("", middleware.RequirePermission("tenant", "create"), h.Create)
		tenants.PUT("/:id", middleware.RequirePermission("tenant", "update"), h.Update)
		tenants.DELETE("/:id", middleware.RequirePermission("tenant", "delete"), h.Delete)
		tenants.GET("/:id", middleware.RequirePermission("tenant", "list"), h.Get)
		tenants.GET("", middleware.RequirePermission("tenant", "list"), h.List)
	}
}

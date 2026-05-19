package organization

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	orgs := protected.Group("/organizations")
	{
		orgs.GET("/tree", middleware.RequirePermission("organization", "list"), h.Tree)
		orgs.GET("", middleware.RequirePermission("organization", "list"), h.List)
		orgs.GET("/:id", middleware.RequirePermission("organization", "list"), h.Get)
		orgs.POST("", middleware.RequirePermission("organization", "create"), h.Create)
		orgs.PUT("/:id", middleware.RequirePermission("organization", "update"), h.Update)
		orgs.DELETE("/:id", middleware.RequirePermission("organization", "delete"), h.Delete)
	}
}

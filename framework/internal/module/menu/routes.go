package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	menus := protected.Group("/menus")
	{
		menus.GET("/tree", middleware.RequirePermission("menu", "list"), h.Tree)
		menus.GET("", middleware.RequirePermission("menu", "list"), h.List)
		menus.GET("/:id", middleware.RequirePermission("menu", "list"), h.Get)
		menus.POST("", middleware.RequirePermission("menu", "create"), h.Create)
		menus.PUT("/:id", middleware.RequirePermission("menu", "update"), h.Update)
		menus.DELETE("/:id", middleware.RequirePermission("menu", "delete"), h.Delete)
	}
}

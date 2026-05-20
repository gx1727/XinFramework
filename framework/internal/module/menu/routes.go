package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	menus := protected.Group("/menus")
	{
		menus.GET("/tree", middleware.RequirePermission(permission.ResMenu, permission.ActList), h.Tree)
		menus.GET("", middleware.RequirePermission(permission.ResMenu, permission.ActList), h.List)
		menus.GET("/:id", middleware.RequirePermission(permission.ResMenu, permission.ActList), h.Get)
		menus.POST("", middleware.RequirePermission(permission.ResMenu, permission.ActCreate), h.Create)
		menus.PUT("/:id", middleware.RequirePermission(permission.ResMenu, permission.ActUpdate), h.Update)
		menus.DELETE("/:id", middleware.RequirePermission(permission.ResMenu, permission.ActDelete), h.Delete)
	}
}

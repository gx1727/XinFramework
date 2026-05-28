package resource

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	resources := protected.Group("/resources")
	{
		resources.GET("", middleware.RequirePermission(permission.ResResource, permission.ActList), h.List)
		resources.GET("/:id", middleware.RequirePermission(permission.ResResource, permission.ActList), h.Get)
		resources.POST("", middleware.RequirePermission(permission.ResResource, permission.ActCreate), h.Create)
		resources.PUT("/:id", middleware.RequirePermission(permission.ResResource, permission.ActUpdate), h.Update)
		resources.DELETE("/:id", middleware.RequirePermission(permission.ResResource, permission.ActDelete), h.Delete)
		resources.GET("/by-menu/:menu_id", middleware.RequirePermission(permission.ResResource, permission.ActList), h.GetByMenu)
		resources.GET("/my", h.GetMyResources)
	}
}

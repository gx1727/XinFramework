package resource

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	resources := protected.Group("/resources")
	{
		resources.GET("", middleware.Require(permission.P(permission.ResResource, permission.ActList)), h.List)
		resources.GET("/:id", middleware.Require(permission.P(permission.ResResource, permission.ActList)), h.Get)
		resources.POST("", middleware.Require(permission.P(permission.ResResource, permission.ActCreate)), h.Create)
		resources.PUT("/:id", middleware.Require(permission.P(permission.ResResource, permission.ActUpdate)), h.Update)
		resources.DELETE("/:id", middleware.Require(permission.P(permission.ResResource, permission.ActDelete)), h.Delete)
		resources.GET("/by-menu/:menu_id", middleware.Require(permission.P(permission.ResResource, permission.ActList)), h.GetByMenu)
		resources.GET("/my", h.GetMyResources)
	}
}

package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(tenant *gin.RouterGroup, h *Handler) {
	menus := tenant.Group("/menus")
	{
		menus.GET("/tree", middleware.Require(permission.P(permission.ResMenu, permission.ActList)), h.Tree)
		menus.GET("", middleware.Require(permission.P(permission.ResMenu, permission.ActList)), h.List)
		menus.GET("/:id", middleware.Require(permission.P(permission.ResMenu, permission.ActList)), h.Get)
		menus.POST("", middleware.Require(permission.P(permission.ResMenu, permission.ActCreate)), h.Create)
		menus.PUT("/:id", middleware.Require(permission.P(permission.ResMenu, permission.ActUpdate)), h.Update)
		menus.DELETE("/:id", middleware.Require(permission.P(permission.ResMenu, permission.ActDelete)), h.Delete)
	}
}

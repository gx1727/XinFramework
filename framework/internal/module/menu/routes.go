package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	menus := protected.Group("/menus")
	{
		menus.POST("", h.Create)
		menus.PUT("/:id", h.Update)
		menus.DELETE("/:id", h.Delete)
		menus.GET("/:id", h.Get)
		menus.GET("", h.List)
		menus.GET("/tree", h.Tree)
	}
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("menu", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

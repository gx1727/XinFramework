package resource

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	resources := protected.Group("/resources")
	{
		resources.GET("", h.List)
		resources.GET("/:id", h.Get)
		resources.POST("", h.Create)
		resources.PUT("/:id", h.Update)
		resources.DELETE("/:id", h.Delete)
		resources.GET("/by-menu/:menu_id", h.GetByMenu)
	}
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("resource", func(_, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

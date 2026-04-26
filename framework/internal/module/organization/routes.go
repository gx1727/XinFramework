package organization

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	orgs := protected.Group("/organizations")
	{
		orgs.GET("", h.List)
		orgs.GET("/:id", h.Get)
		orgs.POST("", h.Create)
		orgs.PUT("/:id", h.Update)
		orgs.DELETE("/:id", h.Delete)
		orgs.GET("/tree", h.Tree)
	}
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("organization", func(_, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

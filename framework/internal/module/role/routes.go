package role

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	roles := protected.Group("/roles")
	{
		roles.GET("", h.List)
		roles.GET("/:id", h.Get)
		roles.POST("", h.Create)
		roles.PUT("/:id", h.Update)
		roles.DELETE("/:id", h.Delete)
		roles.GET("/:id/data-scopes", h.GetDataScopes)
		roles.PUT("/:id/data-scopes", h.UpdateDataScopes)
	}
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("role", func(_, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

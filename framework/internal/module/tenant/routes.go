package tenant

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	tenants := protected.Group("/tenants")
	{
		tenants.POST("", h.Create)
		tenants.PUT("/:id", h.Update)
		tenants.DELETE("/:id", h.Delete)
		tenants.GET("/:id", h.Get)
		tenants.GET("", h.List)
	}
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("tenant", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

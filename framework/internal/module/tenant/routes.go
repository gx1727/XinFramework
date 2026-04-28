package tenant

import (
	"github.com/gin-gonic/gin"
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

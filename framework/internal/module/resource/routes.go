package resource

import (
	"github.com/gin-gonic/gin"
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

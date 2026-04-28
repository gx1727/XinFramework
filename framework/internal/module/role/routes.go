package role

import (
	"github.com/gin-gonic/gin"
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

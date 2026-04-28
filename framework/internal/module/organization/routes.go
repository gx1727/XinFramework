package organization

import (
	"github.com/gin-gonic/gin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	orgs := protected.Group("/organizations")
	{
		orgs.GET("/tree", h.Tree)
		orgs.GET("", h.List)
		orgs.GET("/:id", h.Get)
		orgs.POST("", h.Create)
		orgs.PUT("/:id", h.Update)
		orgs.DELETE("/:id", h.Delete)
	}
}

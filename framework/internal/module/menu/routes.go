package menu

import (
	"github.com/gin-gonic/gin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	menus := protected.Group("/menus")
	{
		menus.GET("/tree", h.Tree)
		menus.GET("", h.List)
		menus.GET("/:id", h.Get)
		menus.POST("", h.Create)
		menus.PUT("/:id", h.Update)
		menus.DELETE("/:id", h.Delete)
	}
}

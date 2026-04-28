package dict

import (
	"github.com/gin-gonic/gin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/dicts", h.List)
	protected.GET("/dicts/:code", h.Get)
	protected.POST("/dicts", h.Create)
	protected.PUT("/dicts/:id", h.Update)
	protected.DELETE("/dicts/:id", h.Delete)

	protected.POST("/dicts/:id/items", h.CreateItem)
	protected.PUT("/dicts/:id/items/:item_id", h.UpdateItem)
	protected.DELETE("/dicts/:id/items/:item_id", h.DeleteItem)
}

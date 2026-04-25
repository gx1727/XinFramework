package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/dicts", h.List)
	protected.GET("/dicts/:code", h.Get)
	protected.POST("/dicts", h.Create)
	protected.PUT("/dicts/:id", h.Update)
	protected.DELETE("/dicts/:id", h.Delete)

	protected.POST("/dicts/:code/items", h.CreateItem)
	protected.PUT("/dicts/:code/items/:item_id", h.UpdateItem)
	protected.DELETE("/dicts/:code/items/:item_id", h.DeleteItem)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("dict", func(_, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

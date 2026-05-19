package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/dicts", middleware.RequirePermission("dict", "list"), h.List)
	protected.GET("/dicts/:code", middleware.RequirePermission("dict", "list"), h.Get)
	protected.POST("/dicts", middleware.RequirePermission("dict", "create"), h.Create)
	protected.PUT("/dicts/:id", middleware.RequirePermission("dict", "update"), h.Update)
	protected.DELETE("/dicts/:id", middleware.RequirePermission("dict", "delete"), h.Delete)

	protected.POST("/dicts/:id/items", middleware.RequirePermission("dict", "update"), h.CreateItem)
	protected.PUT("/dicts/:id/items/:item_id", middleware.RequirePermission("dict", "update"), h.UpdateItem)
	protected.DELETE("/dicts/:id/items/:item_id", middleware.RequirePermission("dict", "update"), h.DeleteItem)
}

package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/dicts", middleware.RequirePermission(permission.ResDict, permission.ActList), h.List)
	protected.GET("/dicts/:code", middleware.RequirePermission(permission.ResDict, permission.ActList), h.Get)
	protected.POST("/dicts", middleware.RequirePermission(permission.ResDict, permission.ActCreate), h.Create)
	protected.PUT("/dicts/:id", middleware.RequirePermission(permission.ResDict, permission.ActUpdate), h.Update)
	protected.DELETE("/dicts/:id", middleware.RequirePermission(permission.ResDict, permission.ActDelete), h.Delete)

	protected.POST("/dicts/:id/items", middleware.RequirePermission(permission.ResDict, permission.ActUpdate), h.CreateItem)
	protected.PUT("/dicts/:id/items/:item_id", middleware.RequirePermission(permission.ResDict, permission.ActUpdate), h.UpdateItem)
	protected.DELETE("/dicts/:id/items/:item_id", middleware.RequirePermission(permission.ResDict, permission.ActUpdate), h.DeleteItem)
}

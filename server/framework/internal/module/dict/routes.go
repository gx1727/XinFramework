package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/dicts", middleware.Require(permission.P(permission.ResDict, permission.ActList)), h.List)
	protected.GET("/dicts/:code", middleware.Require(permission.P(permission.ResDict, permission.ActList)), h.Get)
	protected.POST("/dicts", middleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.Create)
	protected.PUT("/dicts/:id", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.Update)
	protected.DELETE("/dicts/:id", middleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.Delete)

	protected.POST("/dicts/:id/items", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.CreateItem)
	protected.PUT("/dicts/:id/items/:item_id", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpdateItem)
	protected.DELETE("/dicts/:id/items/:item_id", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteItem)
}

// Package dict 字典路由注册
package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	d := protected.Group("/dicts")
	{
		d.GET("", middleware.Require(permission.P(permission.ResDict, permission.ActList)), h.List)
		d.GET("/:id", middleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.Get)
		d.POST("", middleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.Create)
		d.PUT("/:id", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.Update)
		d.DELETE("/:id", middleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.Delete)

		// 字典项：写在同一资源下；权限同上
		d.GET("/:id/items", middleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListItems)
		d.POST("/:id/items", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.CreateItem)
		d.PUT("/:id/items/:item_id", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpdateItem)
		d.DELETE("/:id/items/:item_id", middleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteItem)
	}
}

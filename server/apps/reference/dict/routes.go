// Package dict 字典路由注册
package dict

import (
	"github.com/gin-gonic/gin"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	d := protected.Group("/dicts")
	{
		d.GET("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.List)
		d.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.Get)
		d.POST("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.Create)
		d.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.Update)
		d.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.Delete)

		// 字典项：写在同一资源下；权限同上
		d.GET("/:id/items", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListItems)
		d.POST("/:id/items", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.CreateItem)
		d.PUT("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpdateItem)
		d.DELETE("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteItem)
	}
}

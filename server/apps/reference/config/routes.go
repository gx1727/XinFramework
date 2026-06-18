// Package config 通用配置 - 路由注册
package config

import (
	"github.com/gin-gonic/gin"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	// 公共读（无 auth 也能用，登录页消费 site 信息）
	public.GET("/config", h.GetPublic)

	// 管理端
	g := protected.Group("/config")
	{
		// 分组
		g.GET("/groups", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), h.ListGroups)
		g.POST("/groups", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActCreate)), h.CreateGroup)
		g.PUT("/groups/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), h.UpdateGroup)
		g.DELETE("/groups/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActDelete)), h.DeleteGroup)

		// 分组下的项
		g.GET("/groups/:id/items", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), h.ListItemsByGroup)
		g.POST("/groups/:id/items", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActCreate)), h.CreateItem)

		// 所有项
		g.GET("/items", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), h.ListAllItems)

		// 单项操作
		g.PUT("/items/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), h.UpdateItem)
		g.POST("/items/:id/reset", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), h.ResetItem)
		g.DELETE("/items/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActDelete)), h.DeleteItem)
	}
}

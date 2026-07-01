package task

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// Register 把 task 管理 API 挂在 sys 域。
//
// 路由前缀：/api/v1/sys/tasks/*
// 鉴权：RequireSysRole("super_admin")
//
// 用法：
//
//	protected := slots.MustGet(plugin.SlotProtected).Group
//	task.Register(protected, h)
func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/tasks")
	g.Use(middleware.RequireSysRole("super_admin"))
	{
		g.GET("", middleware.Require(permission.P("task", "list")), h.List)
		g.GET("/stats", middleware.Require(permission.P("task", "list")), h.Stats)
		g.GET("/:id", middleware.Require(permission.P("task", "list")), h.Get)
		g.POST("/:id/cancel", middleware.Require(permission.P("task", "delete")), h.Cancel)
		g.POST("/:id/requeue", middleware.Require(permission.P("task", "update")), h.Requeue)
		g.POST("/cleanup", middleware.Require(permission.P("task", "delete")), h.Cleanup)
	}
}

package task

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// RegisterCron 把 cron job 管理 API 挂在 sys 域。
//
// 路由前缀：/api/v1/sys/cron-jobs/*
// 鉴权：RequireSysRole("super_admin")
//
// 用法：
//
//	protected := slots.MustGet(plugin.SlotProtected).Group
//	task.RegisterCron(protected, h)
func RegisterCron(protected *gin.RouterGroup, h *CronHandler) {
	g := protected.Group("/cron-jobs")
	g.Use(middleware.RequireSysRole("super_admin"))
	{
		g.GET("", middleware.Require(permission.P("task", "list")), h.List)
		g.GET("/:name", middleware.Require(permission.P("task", "list")), h.Get)
		g.POST("", middleware.Require(permission.P("task", "create")), h.Create)
		g.PUT("/:name", middleware.Require(permission.P("task", "update")), h.Update)
		g.DELETE("/:name", middleware.Require(permission.P("task", "delete")), h.Delete)
		g.POST("/:name/enable", middleware.Require(permission.P("task", "update")), h.Enable)
		g.POST("/:name/disable", middleware.Require(permission.P("task", "update")), h.Disable)
		g.POST("/:name/trigger", middleware.Require(permission.P("task", "update")), h.Trigger)
	}
}

package message

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// Register 把站内信路由挂到 tenant 域（/api/v1/messages）。
//
// 必须挂在 tenant *gin.RouterGroup 上 —— 上一级已经过 Auth + RequireTenantContext，
// 这里的 Require(...) 只校验资源码 message.{action}。
func Register(tenant *gin.RouterGroup, h *Handler) {
	g := tenant.Group("/messages")
	{
		// 列表（inbox / sent）
		g.GET("", middleware.Require(permission.P(permission.ResMessage, permission.ActList)), h.List)

		// 单条详情
		g.GET("/:id", middleware.Require(permission.P(permission.ResMessage, permission.ActList)), h.Get)

		// 发送
		g.POST("", middleware.Require(permission.P(permission.ResMessage, permission.ActCreate)), h.Send)

		// 更新（仅发件人可改自己发出去的信，service 层兜底）
		g.PATCH("/:id", middleware.Require(permission.P(permission.ResMessage, permission.ActUpdate)), h.Update)

		// 标记已读
		g.PATCH("/:id/read", middleware.Require(permission.P(permission.ResMessage, permission.ActUpdate)), h.MarkRead)

		// 删除
		g.DELETE("/:id", middleware.Require(permission.P(permission.ResMessage, permission.ActDelete)), h.Delete)

		// 未读数（badge）
		g.GET("/unread-count", middleware.Require(permission.P(permission.ResMessage, permission.ActList)), h.UnreadCount)
	}
}
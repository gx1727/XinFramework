package weixin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.GET("/weixin/ping", func(c *gin.Context) {
		// Health check
	})

	// 小程序登录
	public.POST("/weixin/login", h.Login)
	public.POST("/weixin/phone", h.GetPhoneNumber)

	// 需要登录的接口
	protected.POST("/weixin/bind-phone", middleware.RequireAuthenticated(), h.BindPhone)
}

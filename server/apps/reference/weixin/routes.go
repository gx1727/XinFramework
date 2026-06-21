package weixin

import (
	"github.com/gin-gonic/gin"
)

func Register(public *gin.RouterGroup, tenant *gin.RouterGroup, h *Handler) {
	public.GET("/weixin/ping", func(c *gin.Context) {
		// Health check
	})

	// 小程序登录
	public.POST("/weixin/login", h.Login)
	public.POST("/weixin/phone", h.GetPhoneNumber)

	// 需要登录的接口
	tenant.POST("/weixin/bind-phone", h.BindPhone)
}

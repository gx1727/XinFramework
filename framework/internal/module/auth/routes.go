package auth

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.POST("/auth/login", h.Login)
	public.POST("/auth/register", h.Register)
	public.POST("/auth/refresh", h.Refresh)

	protected.POST("/auth/logout", h.Logout)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("auth", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		Register(public, protected, h)
	})
}

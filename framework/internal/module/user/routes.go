package user

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.POST("/login", h.Login)
	public.POST("/register", h.Register)
	protected.POST("/logout", h.Logout)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("user", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		Register(public, protected, h)
	})
}

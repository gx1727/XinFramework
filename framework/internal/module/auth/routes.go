package auth

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func RegisterV1(public *gin.RouterGroup, protected *gin.RouterGroup) {
	h := NewHandler()
	public.POST("/login", h.Login)
	public.POST("/register", h.Register)
	protected.POST("/logout", h.Logout)
}

func Module() plugin.Module {
	return plugin.NewModule("auth", RegisterV1)
}

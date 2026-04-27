package user

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/users", h.List)
	protected.GET("/users/:id", h.Get)
	protected.PUT("/users/:id/status", h.UpdateStatus)
	protected.GET("/user/profile", h.Profile)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("user", func(_, protected *gin.RouterGroup) {
		Register(protected, h)
	})
}

package weixin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module(h *Handler) plugin.Module {
	return plugin.NewModule("weixin", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		Register(public, protected, h)
	})
}

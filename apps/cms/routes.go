package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	protected.GET("/cms/ping", h.Ping)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModuleWithOpts("cms",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			Register(public, protected, h)
		},
		plugin.WithInit(InitConfig),
	)
}

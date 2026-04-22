package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
)

func RegisterV1(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})
}

func Module() plugin.Module {
	return plugin.NewModule("system", RegisterV1)
}

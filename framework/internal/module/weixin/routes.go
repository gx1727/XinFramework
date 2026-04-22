package weixin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	protected.GET("/weixin/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{"domain": "weixin", "status": "enabled"})
	})
}

func Module() plugin.Module {
	return plugin.NewModule("weixin", Register)
}

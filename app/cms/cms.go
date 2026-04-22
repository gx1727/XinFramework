package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
)

func RegisterV1(public, protected *gin.RouterGroup) {
	protected.GET("/cms/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{"domain": "cms", "status": "enabled"})
	})
}

func Module() plugin.Module {
	return plugin.NewModule("cms", RegisterV1)
}

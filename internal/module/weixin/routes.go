package weixin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/pkg/resp"
)

// RegisterV1 registers weixin routes for v1.
func RegisterV1(r *gin.RouterGroup) {
	r.GET("/weixin/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{"domain": "weixin", "status": "enabled"})
	})
}

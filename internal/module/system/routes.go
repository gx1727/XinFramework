package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/pkg/resp"
)

func RegisterV1(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})
}

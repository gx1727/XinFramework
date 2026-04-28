package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/resp"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})
}

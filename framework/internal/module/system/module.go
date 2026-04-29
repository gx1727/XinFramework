package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
)

// Module 返回 system 模块的完整定义（不需要 app 参数）
func Module() plugin.Module {
	return plugin.NewModule("system", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		public.GET("/health", func(c *gin.Context) {
			resp.Success(c, gin.H{"status": "ok"})
		})
	})
}

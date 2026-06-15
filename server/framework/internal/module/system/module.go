package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module 返回 system 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("system", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler()
		Register(public, protected, h)
	})
}

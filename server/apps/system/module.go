package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 system 模块的完整定义
//
// Phase 5：显式接收 *appx.App。system 不需要 DB/config 但签名仍统一。
func Module(_ *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "system",
		RegFn: func(_ plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			h := NewHandler()
			Register(public, protected, h)
		},
	}
}

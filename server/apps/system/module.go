package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 system 模块的完整定义
//
func Module(_ *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "system",
		RegFn: func(_ plugin.Reader, public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			h := NewHandler()
			Register(public, tenant, protected, h)
		},
	}
}

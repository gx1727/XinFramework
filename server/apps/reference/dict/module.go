// Package dict 字典模块入口
package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 dict 模块的完整定义
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "dict",
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			h := NewHandler(NewService(pool, NewPostgresDictRepository(pool)))
			Register(tenant, protected, h)
		},
	}
}

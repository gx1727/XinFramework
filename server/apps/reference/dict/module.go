// Package dict 字典模块入口
package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module 返回 dict 模块的完整定义
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "dict",
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			h := NewHandler(NewService(pool, NewPostgresDictRepository(pool)))
			Register(protected, h)
		},
	}
}

package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module 返回 menu 模块的完整定义
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "menu",
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			h := NewHandler(NewService(NewMenuRepository(pool)))
			Register(protected, h)
		},
	}
}

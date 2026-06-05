package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 menu 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("menu", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(NewMenuRepository(db.Get())))
		Register(protected, h)
	})
}

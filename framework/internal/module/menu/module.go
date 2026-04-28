package menu

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 menu 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("menu", func(public, protected *gin.RouterGroup) {
		h := NewHandler(NewService(app.Repository.Menu()))
		Register(protected, h)
	})
}

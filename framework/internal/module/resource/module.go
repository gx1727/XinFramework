package resource

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 resource 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("resource", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(app.Repository.Resource(), app.Repository.Menu()))
		Register(protected, h)
	})
}

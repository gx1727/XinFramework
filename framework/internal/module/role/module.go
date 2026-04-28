package role

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 role 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("role", func(public, protected *gin.RouterGroup) {
		h := NewHandler(NewService(app.Repository.Role(), app.Repository.DataScope()))
		Register(protected, h)
	})
}

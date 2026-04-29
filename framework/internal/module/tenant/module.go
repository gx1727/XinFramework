package tenant

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 tenant 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("tenant", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(app.Repository.Tenant()))
		Register(protected, h)
	})
}

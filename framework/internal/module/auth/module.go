package auth

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 auth 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("auth", func(public, protected *gin.RouterGroup) {
		repos := Repositories{
			Account: app.Repository.Account(),
			Tenant:  app.Repository.Tenant(),
			Role:    app.Repository.Role(),
			User:    app.Repository.User(),
		}
		deps := DefaultDependencies(app.Config, app.DB, repos)
		h := NewHandler(NewService(deps))
		Register(public, protected, h)
	})
}

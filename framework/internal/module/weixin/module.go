package weixin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 weixin 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModuleWithOpts("weixin",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			svc := NewService(
				app.DB,
				app.SessionMgr,
				app.Repository.AccountAuth(),
				app.Repository.Account(),
				app.Repository.Tenant(),
				app.Repository.Role(),
				app.Repository.User(),
			)
			h := NewHandler(svc)
			Register(public, protected, h)
		},
		plugin.WithInit(func() error {
			return InitConfig()
		}),
	)
}

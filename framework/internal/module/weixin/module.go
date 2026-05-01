package weixin

import (
	"gx1727.com/xin/framework/internal/module/tenant"

	"gx1727.com/xin/framework/internal/module/auth"

	"gx1727.com/xin/framework/internal/module/role"

	"gx1727.com/xin/framework/internal/module/user"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

// Module 返回 weixin 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModuleWithOpts("weixin",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			svc := NewService(
				db.Get(),
				session.Manager(),
				auth.NewAccountAuthRepository(db.Get()),
				auth.NewAccountRepository(db.Get()),
				tenant.NewTenantRepository(db.Get()),
				role.NewRoleRepository(db.Get()),
				user.NewUserRepository(db.Get()),
			)
			h := NewHandler(svc)
			Register(public, protected, h)
		},
		plugin.WithInit(func() error {
			return InitConfig()
		}),
	)
}

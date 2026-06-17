package auth

import (
	"gx1727.com/xin/framework/pkg/config"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/apps/boot/tenant"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())

	// Phase 2: register this module's AccountRepository factory with
	// the framework's pkg/auth registry so that framework/internal
	// modules (currently user) can resolve account access at runtime.
	//
	// Phase 3 will retire this indirection: once user moves to
	// apps/rbac/user/, it can import apps/boot/auth directly.
	pkgauth.Register(func() pkgauth.AccountRepository {
		return NewAccountRepository(db.Get())
	})

	// 同样注册 AccountAuthRepository，让 apps/reference/weixin 这类下游
	// 模块在不直接 import apps/boot/auth 的前提下拿到第三方授权（微信/QQ）
	// 绑定的数据访问实现。
	pkgauth.RegisterAccountAuthRepository(func() pkgauth.AccountAuthRepository {
		return NewAccountAuthRepository(db.Get())
	})
}

func Module() plugin.Module {
	return plugin.NewModule("auth", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		repos := Repositories{
			Account:  NewAccountRepository(db.Get()),
			Tenant:   tenant.NewTenantRepository(db.Get()),
			Platform: permission.NewPlatformRoleRepository(db.Get()),
		}
		deps := DefaultDependencies(config.Get(), db.Get(), repos)
		h := NewHandler(NewService(deps))
		Register(public, protected, h)
	})
}

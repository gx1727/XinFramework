package weixin

import (
	"github.com/gin-gonic/gin"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
	"gx1727.com/xin/framework/pkg/session"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

func init() {
	plugin.Register(Module())
}

// Module returns the weixin module.
//
// weixin now lives under apps/reference/weixin (Phase 3b). It depends
// on apps/boot/{auth,tenant} and apps/rbac/{user,role} — but instead
// of importing them directly, it goes through the public pkg/{auth,
// tenant, rbac} registries. This keeps weixin's compile-time imports
// limited to the framework module, matching the convention used by
// every other app.
func Module() plugin.Module {
	return plugin.NewModuleWithOpts("weixin",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			accountAuthFactory := pkgauth.GetAccountAuthRepository()
			accountFactory := pkgauth.Get()
			tenantFactory := pkgtenant.Get()
			roleFactory := pkgrbac.GetRoleRepository()
			userFactory := pkgrbac.GetUserRepository()

			if accountAuthFactory == nil || accountFactory == nil ||
				tenantFactory == nil || roleFactory == nil || userFactory == nil {
				// 必装模块未加载：拒绝注册路由
				return
			}

			svc := NewService(
				db.Get(),
				session.Manager(),
				accountAuthFactory(),
				accountFactory(),
				tenantFactory(),
				roleFactory(),
				userFactory(),
			)
			h := NewHandler(svc)
			Register(public, protected, h)
		},
		plugin.WithInit(func() error {
			return InitConfig()
		}),
	)
}

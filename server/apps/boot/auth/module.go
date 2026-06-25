package auth

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

// Module returns the auth module as a BaseModule.
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "auth",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB
			w.SetAccountRepo(NewAccountRepository(pool))
			w.SetAccountAuthRepo(NewAccountAuthRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, tenantGroup *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}

			// 跨模块依赖（tenant repo）从 AppContext 拿，遵循 DI 设计意图。
			// auth 与 tenants 都是 alwaysOn 模块，理论 ctx.TenantRepo() 必非 nil；
			// 若为 nil（启动顺序异常），拒绝注册路由，避免后续 panic。
			tenantRepo := pkgtenant.TenantRepository(nil)
			if ctx != nil {
				tenantRepo = ctx.TenantRepo()
			}
			if tenantRepo == nil {
				return
			}

			repos := Repositories{
				Account:  NewAccountRepository(pool),
				Tenant:   tenantRepo,
				Platform: permission.NewPlatformRoleRepository(pool),
			}
			deps := DefaultDependencies(app.Config, pool, repos)
			h := NewHandler(NewService(deps))
			Register(public, protected, h)
		},
	}
}

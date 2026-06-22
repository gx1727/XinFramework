package auth

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/apps/platform/tenant"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the auth module as a BaseModule.
//
// 也不再通过 init() 自动注册，main.go 显式调用 auth.Module(app) 即可。
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

			tenantRepo := tenant.NewTenantRepository(pool)
			if ctx != nil {
				if tr := ctx.TenantRepo(); tr != nil {
					_ = tr
				}
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

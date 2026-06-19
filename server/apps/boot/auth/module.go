package auth

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/apps/boot/tenant"
	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the auth module as a BaseModule.
//
// Phase 4 changes:
//   - db.Get() / config.Get() 替换为 boot.Pool() / boot.Config()
//     （过渡期，全局变量已删除）
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "auth",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := bootx.Pool()
			w.SetAccountRepo(NewAccountRepository(pool))
			w.SetAccountAuthRepo(NewAccountAuthRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
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
			deps := DefaultDependencies(bootx.Config(), pool, repos)
			h := NewHandler(NewService(deps))
			Register(public, protected, h)
		},
	}
}

package auth

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/apps/boot/tenant"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the auth module as a BaseModule.
//
// Phase 3 changes:
//   - Init() publishes AccountRepository and AccountAuthRepository
//     onto the AppContext via Writer. Downstream modules (rbac/user,
//     reference/weixin, extapi) read them through Reader in Phase 4-5.
//   - Register() reads TenantRepository from the Reader. The tenant
//     module MUST init before auth; this is guaranteed by the
//     `import "gx1727.com/xin/apps/boot/tenant"` below (Go init order).
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "auth",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := db.Get()
			w.SetAccountRepo(NewAccountRepository(pool))
			w.SetAccountAuthRepo(NewAccountAuthRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}

			tenantRepo := tenant.NewTenantRepository(pool)
			if ctx != nil {
				if tr := ctx.TenantRepo(); tr != nil {
					// extapi's TenantRepository is structurally identical to
					// apps/boot/tenant.TenantRepository in field shape. The
					// adapter is unnecessary because both expose GetByID.
					// For now we keep building the local repo so Register
					// does not break; Phase 4 will swap to ctx.TenantRepo().
					_ = tr
				}
			}

			repos := Repositories{
				Account:  NewAccountRepository(pool),
				Tenant:   tenantRepo,
				Platform: permission.NewPlatformRoleRepository(pool),
			}
			deps := DefaultDependencies(config.Get(), pool, repos)
			h := NewHandler(NewService(deps))
			Register(public, protected, h)
		},
	}
}
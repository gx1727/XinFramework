package weixin

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

func init() {
	plugin.Register(Module())
}

// Module returns the weixin module as a BaseModule.
//
// weixin depends on apps/boot/{auth,tenant} and apps/rbac/{user,role}.
// Phase 3 changes the lookup from the legacy pkgauth.Get/pkgtenant.Get
// globals to AppContext.Reader. The Init phase runs once at boot and
// calls InitConfig(); downstream dependencies are resolved lazily on
// first request through the closed-over reader.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "weixin",
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			// Phase 4: 通过 bootx 拿全局 config，注入 weixinCfg；service 不再直接调 config.Get()。
			SetGlobalConfig(bootx.Config())
			return InitConfig()
		},
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			accountRepo := ctx.AccountRepo()
			accountAuthRepo := ctx.AccountAuthRepo()
			tenantRepo := ctx.TenantRepo()
			roleRepo := ctx.RoleRepo()
			userRepo := ctx.UserRepo()

			if accountRepo == nil || accountAuthRepo == nil ||
				tenantRepo == nil || roleRepo == nil || userRepo == nil {
				// Required modules not loaded: refuse to register routes.
				return
			}

			svc := NewService(
				bootx.Pool(),
				session.Manager(),
				accountAuthRepo,
				accountRepo,
				tenantRepo,
				roleRepo,
				userRepo,
			)
			h := NewHandler(svc)
			Register(public, protected, h)
		},
	}
}
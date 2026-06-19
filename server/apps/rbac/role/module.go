package role

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the role module as a BaseModule.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "role",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := bootx.Pool()
			w.SetRoleRepo(NewRoleRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(
				pool,
				NewRoleRepository(pool),
				permission.NewDataScopeRepository(pool),
				NewRoleMenuRepository(pool),
				authzSvc,
			))
			Register(protected, h)
		},
	}
}

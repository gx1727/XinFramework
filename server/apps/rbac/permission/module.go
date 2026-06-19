package permission

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the permission module as a BaseModule.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "permission",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := bootx.Pool()
			w.SetPermRepo(NewRoleResourceRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			roleResourceRepo := NewRoleResourceRepository(pool)
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(pool, roleResourceRepo, authzSvc))
			Register(protected, h)
		},
	}
}

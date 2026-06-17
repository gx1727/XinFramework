package permission

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the permission module as a BaseModule.
//
// Phase 4: publishes RoleResourceRepository onto the AppContext.Writer.
// The Auth middleware (framework/pkg/middleware) consumes it via
// ctx.PermRepo() when resolving effective permissions.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "permission",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := db.Get()
			w.SetPermRepo(NewRoleResourceRepository(pool))
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			roleResourceRepo := NewRoleResourceRepository(pool)
			h := NewHandler(NewService(pool, roleResourceRepo))
			Register(protected, h)
		},
	}
}
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
//
// Phase 5 Step 7: pulls Authorization from AppContext.Reader to
// invalidate user permission cache after role-resource changes.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "permission",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := db.Get()
			w.SetPermRepo(NewRoleResourceRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			roleResourceRepo := NewRoleResourceRepository(pool)
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(pool, roleResourceRepo, authzSvc))
			Register(protected, h)
		},
	}
}
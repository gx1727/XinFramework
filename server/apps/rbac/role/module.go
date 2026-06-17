package role

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the role module as a BaseModule.
//
// Phase 4: publishes RoleRepository onto the AppContext.Writer.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "role",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := db.Get()
			w.SetRoleRepo(NewRoleRepository(pool))
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			h := NewHandler(NewService(
				NewRoleRepository(pool),
				permission.NewDataScopeRepository(pool),
				NewRoleMenuRepository(pool),
			))
			Register(protected, h)
		},
	}
}
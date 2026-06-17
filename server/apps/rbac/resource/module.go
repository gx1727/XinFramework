package resource

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the resource module as a BaseModule.
//
// Phase 5 Step 6: pulls Authorization from AppContext.Reader to
// invalidate user permission cache after resource changes.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "resource",
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(NewResourceRepository(pool), authzSvc))
			Register(protected, h)
		},
	}
}
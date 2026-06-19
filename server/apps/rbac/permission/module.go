package permission

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the permission module as a BaseModule.
//
// Phase 5：显式接收 *appx.App。
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "permission",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB
			w.SetPermRepo(NewRoleResourceRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			roleResourceRepo := NewRoleResourceRepository(pool)
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(pool, roleResourceRepo, authzSvc))
			Register(protected, h)
		},
	}
}

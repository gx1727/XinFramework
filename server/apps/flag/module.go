package flag

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the flag module as a BaseModule.
//
// Phase 5：显式接收 *appx.App。
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "flag",
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}
			InitRepositories(pool)
			SetConfig(app.Config)

			h := NewHandler()
			Register(public, protected, h)
		},
	}
}

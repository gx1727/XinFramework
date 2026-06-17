package flag

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the flag module as a BaseModule.
//
// Migration note (Phase 5): replace the `db.Get()` fallback with
// `ctx.DB()` once every module is wired against AppContext.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "flag",
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}
			InitRepositories(pool)

			h := NewHandler()
			Register(public, protected, h)
		},
	}
}
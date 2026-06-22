package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the cms module as a BaseModule.
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "cms",
		RegFn: func(ctx plugin.Reader, public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			h := NewHandler(app.DB, ctx)
			Register(h, public, protected)
		},
	}
}

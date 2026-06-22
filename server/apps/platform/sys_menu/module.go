package sysmenu

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "sys_menu",
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			h := NewHandler(NewService(pool, NewRepository(pool)))
			Register(protected, h)
		},
	}
}

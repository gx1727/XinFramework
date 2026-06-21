package platformmenu

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 platform_menu 模块定义。
//
// Phase 5 风格：显式接收 *appx.App，自己从 app.DB() 拿 pool，
// 自己构造 repository/service/handler。零全局变量。
//
// 模块名约定："platform_menu"（与 apps/rbac/menu 的 "menu" 区分）。
// 在 cfg.Module: 里以 "platform_menu" 标识。
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "platform_menu",
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			h := NewHandler(NewService(pool, NewMenuRepository(pool)))
			Register(protected, h)
		},
	}
}

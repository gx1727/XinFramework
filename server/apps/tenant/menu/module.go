package menu

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 menu 模块的完整定义
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "menu",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB.Raw()
			h := NewHandler(NewService(NewMenuRepository(pool)))
			Register(tenant, h)
		},
	}
}

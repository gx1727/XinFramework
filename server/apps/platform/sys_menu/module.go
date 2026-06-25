package sysmenu

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "sys_menu",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			protected := slots.MustGet(plugin.SlotProtected).Group
			pool := app.DB
			h := NewHandler(NewService(pool, NewRepository(pool)))
			Register(protected, h)
		},
	}
}

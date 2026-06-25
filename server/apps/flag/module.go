package flag

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the flag module as a BaseModule.
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "flag",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}
			h := NewHandler(pool, app.Config)
			Register(public, tenant, h)
		},
	}
}

package resource

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the resource module as a BaseModule.
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "resource",
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(NewResourceRepository(pool), authzSvc))
			Register(tenant, h)
		},
	}
}

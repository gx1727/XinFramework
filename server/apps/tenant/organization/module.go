package organization

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the organization module as a BaseModule.
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "organization",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB
			w.SetOrgRepo(NewOrganizationRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB
			h := NewHandler(NewService(pool, NewOrganizationRepository(pool)))
			Register(tenant, h)
		},
	}
}

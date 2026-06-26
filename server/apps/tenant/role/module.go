package role

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the role module as a BaseModule.
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "role",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB.Raw()
			w.SetRoleRepo(NewRoleRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB.Raw()
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(
				pool,
				NewRoleRepository(pool),
				permission.NewDataScopeRepository(pool),
				NewRoleMenuRepository(pool),
				authzSvc,
			))
			Register(tenant, h)
		},
	}
}

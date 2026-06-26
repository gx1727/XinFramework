package permission

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the permission module as a BaseModule.
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "permission",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB.Raw()
			w.SetPermRepo(NewRoleResourceRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB.Raw()
			roleResourceRepo := NewRoleResourceRepository(pool)
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(pool, roleResourceRepo, authzSvc))
			Register(tenant, h)
		},
	}
}

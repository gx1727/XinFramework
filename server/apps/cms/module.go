package cms

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module returns the cms module as a BaseModule.
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "cms",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
			h := NewHandler(app.DB.Raw(), ctx)
			Register(h, public, protected)
		},
	}
}

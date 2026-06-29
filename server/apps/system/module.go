package system

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 system 模块的完整定义
//
func Module(_ *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "system",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			tenant := slots.MustGet(plugin.SlotTenant).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
			h := NewHandler()
			Register(public, tenant, protected, h)
		},
	}
}

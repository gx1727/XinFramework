// Package dict 字典模块入口
package dict

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 dict 模块的完整定义
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "dict",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
			pool := app.DB.Raw()
			h := NewHandler(NewService(pool, NewPostgresDictRepository(pool)))
			Register(tenant, protected, h)
		},
	}
}

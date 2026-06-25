package sysuser

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 sys_user 模块定义。
//
// 自己构造 repository/service/handler。零全局变量。
//
// 模块名约定："sys_user"。在 cfg.Module: 里以 "sys_user" 标识。
// alwaysOn 列表中：phase 0023+ 阶段将 sys_user 加入 alwaysOn
// （平台管理员身份管理是 platform 域核心，必须常驻）。
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "sys_user",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			protected := slots.MustGet(plugin.SlotProtected).Group
			pool := app.DB
			h := NewHandler(NewService(pool, NewRepository(pool)))
			Register(protected, h)
		},
	}
}

package sysuser

import (
	"gx1727.com/xin/framework/pkg/appx"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
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
			pool := app.DB.Raw()
			// accountRepo 由 apps/boot/auth 在 Init 阶段注入。
			// 未注入时 service 仍可工作（只能使用模式 1 绑定已有账号），
			// 仅模式 2（一并建可登录账号）会报 ErrSysUserAccountRepoMissing。
			var accountRepo pkgauth.AccountRepository
			if ctx != nil {
				accountRepo = ctx.AccountRepo()
			}
			h := NewHandler(NewService(pool, NewRepository(pool), accountRepo))
			Register(protected, h)
		},
	}
}

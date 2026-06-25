package message

import (
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回站内信模块。
//
// 接入：
//   1. cmd/xin/main.go：显式 import 并加进 []plugin.Module
//   2. framework/pkg/config/config.go optOutModules 加 "message"
//   3. framework/pkg/permission/constants.go 加 ResMessage = "message"
//   4. framework/pkg/resp/errors.go 加 CodeMessage = 16000
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "message",
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			tenant := slots.MustGet(plugin.SlotTenant).Group
			pool := app.DB
			repo := NewRepository(pool)
			svc := NewService(pool, repo)
			h := NewHandler(svc)

			// 业务域路由（Auth + RequireTenantContext 已在 tenant 上游完成）
			Register(tenant, h)
		},
	}
}
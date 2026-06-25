// Package config 通用配置模块入口
//
//   - 三 handler 拆分（Business / Platform / Public）
//   - 路由 /configs 业务 + /configs/platform 平台 + /configs/public 公共
//   - Resolve / Override / Visibility 三大业务能力
package config

import (
	"context"
	"log"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 config 模块的完整定义
func Module(app *appx.App) plugin.Module {
	pool := app.DB
	repo := NewPostgresConfigRepository(pool)
	cache := NewCache()
	svc := NewService(pool, repo, cache)

	return &plugin.BaseModule{
		NameStr: "config",

		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := EnsureTemplateSeeded(ctx, pool); err != nil {
				log.Printf("[config] init self-check seed skipped: %v", err)
			}
			if err := HealConfigMenuParent(ctx, pool); err != nil {
				log.Printf("[config] heal config menu parent skipped: %v", err)
			}
			return nil
		},

		// RegFn: 注册三组路由（业务 + 平台 + 公共）
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			tenant := slots.MustGet(plugin.SlotTenant).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
			bh := NewBusinessHandler(svc)
			ph := NewPlatformHandler(svc)
			pubh := NewPublicHandler(svc)
			Register(public, tenant, protected, bh, ph, pubh)
		},
	}
}

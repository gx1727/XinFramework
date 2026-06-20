// Package config 通用配置模块入口
package config

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 config 模块的完整定义
//
// Phase 5：显式接收 *appx.App。
func Module(app *appx.App) plugin.Module {
	pool := app.DB
	repo := NewPostgresConfigRepository(pool)
	cache := NewCache()
	svc := NewService(pool, repo, cache)
	h := NewHandler(svc)

	return &plugin.BaseModule{
		NameStr: "config",

		// InitFn: 启动期自检 bootstrap 是否有 config 数据，没有就补 seed。
		// 解决"老 framework.sql 部署过的库"在新 framework.sql 加了 config seed
		// 但因 _schema_migrations 已标记 framework.sql 而跳过导致的缺口。
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pool := app.DB

			// 1) seed bootstrap
			if err := EnsureTemplateSeeded(ctx, pool); err != nil {
				log.Printf("[config] init self-check seed skipped: %v", err)
			}

			// 2) 自愈：修复老 config menu 的 parent_id（写死 5 导致的孤儿）
			if err := HealConfigMenuParent(ctx, pool); err != nil {
				log.Printf("[config] heal config menu parent skipped: %v", err)
			}

			return nil
		},

		RegFn: func(_ plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			Register(public, protected, h)
		},
	}
}

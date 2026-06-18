// Package config 通用配置模块入口
package config

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module 返回 config 模块的完整定义
func Module() plugin.Module {
	repo := NewPostgresConfigRepository(db.Get())
	cache := NewCache()
	svc := NewService(repo, cache)
	h := NewHandler(svc)

	return &plugin.BaseModule{
		NameStr: "config",

		// InitFn: 启动期自检 __template__ 是否有 config 数据，没有就补 seed。
		// 解决"老 framework.sql 部署过的库"在新 framework.sql 加了 config seed
		// 但因 _schema_migrations 已标记 framework.sql 而跳过导致的缺口。
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := EnsureTemplateSeeded(ctx); err != nil {
				// 自检失败不阻塞启动（seed 失败也无所谓，业务表可能还没建）
				log.Printf("[config] init self-check seed skipped: %v", err)
			}
			return nil
		},

		RegFn: func(_ plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			Register(public, protected, h)
		},
	}
}

// Package boot 启动编排：构造 App，把基础设施（DB、Config、Session、Authz）
// 装进 *appx.App 并暴露给后续阶段。
//
// Phase 5 改动：
//   - 完全删除 globalApp / Pool() / Config() 等过渡期访问器
//   - Init 直接返回 *appx.App，让 main.go 显式传给每个模块
//   - 不再有任何包级可变状态（除 logger/cache 这类自有生命周期的子系统）
package boot

import (
	"context"
	"fmt"
	"log"

	"gx1727.com/xin/framework/internal/core/ext_impl"
	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/internal/service"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/authz"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/dict"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

// Init 构造 *appx.App：打开数据库、加载配置、初始化缓存、session、权限服务。
// 进程级资源都装进返回的 App，调用方负责显式传给模块。
func Init(cfg *config.Config) (*appx.App, error) {
	logger.Init(cfg.Log.Dir, cfg.Log.Level)

	ctx := context.Background()
	pool, err := db.Init(ctx, &cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}

	dict.Init(pool)

	if err := cache.Init(&cfg.Redis); err != nil {
		pool.Close()
		return nil, fmt.Errorf("cache init failed: %w", err)
	}

	var sm session.SessionManager
	if cache.Get() != nil {
		sm = session.NewRedisSessionManager()
	} else {
		sm = session.NewDBSessionManager(pool)
	}
	session.Init(sm)

	var permCache permission.PermissionCache
	if cache.Get() != nil {
		permCache = permission.NewRedisPermissionCache()
	}

	appCtx := plugin.NewAppContext(pool, cache.Get(), cfg, sm)
	ext_impl.InitExtApi(appCtx)

	permService := service.NewPermissionService(
		permission.NewPermissionRepository(pool),
		permission.NewDataScopeRepository(pool),
		permCache,
		permission.NewPlatformRoleRepository(pool),
	)
	authzService := service.NewAuthorizationService(permService)
	authzSvc := authz.Wrap(authzService)
	appCtx.SetAuthz(authzSvc)

	app := &appx.App{
		Config:      cfg,
		DB:          pool,
		SessionMgr:  sm,
		Server:      server.New(cfg),
		PermService: permService,
		Authz:       authzService,
		AppContext:  appCtx,
	}
	_ = authzSvc // kept for future use; currently only AppContext.Authz is read

	// 启动期引导
	if bcfg := LoadBootstrapConfig(); bcfg.Enabled {
		if err := RunBootstrap(ctx, pool, bcfg); err != nil {
			log.Printf("[bootstrap] failed: %v", err)
		}
	}

	return app, nil
}

// Shutdown 释放资源。
func Shutdown(app *appx.App) {
	if app == nil {
		return
	}
	var reader plugin.Reader
	for _, m := range plugin.Apps() {
		if err := m.Shutdown(reader); err != nil {
			log.Printf("module %s shutdown failed: %v", m.Name(), err)
		}
	}
	if err := cache.Close(); err != nil {
		log.Printf("cache close failed: %v", err)
	}
	if app.DB != nil {
		app.DB.Close()
	}
	logger.Close()
}

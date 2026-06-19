// Package boot 启动编排：构造 App，把基础设施（DB、Config、Session、Authz）
// 装进 App 并暴露给后续阶段。
//
// Phase 4 改动：
//   - 不再使用 db.Pool / config.cfg 等包级全局
//   - db.Init() 返回 pool 显式持有
//   - 暴露 boot.Pool() / boot.Config() 给模块使用（过渡期）
//   - 这些 accessor 在未来 main.go 显式构造后会被删除
package boot

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/core/ext_impl"
	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/internal/service"
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

// App 是启动后所有进程级资源的容器，由 Init 构造、Shutdown 释放。
type App struct {
	Config      *config.Config
	DB          *pgxpool.Pool
	SessionMgr  session.SessionManager
	Server      *server.XinServer
	PermService *service.PermissionService
	Authz       *service.AuthorizationService
	AppContext  *plugin.AppContext
}

// globalApp 是过渡期 boot.Pool() / boot.Config() 的来源。
//
// 进程级单例只有一个 App，所以保留一个 package-level 指针是合理的。
// 未来 main.go 显式构造后此变量 + Pool()/Config() 整体删除，
// 届时每个模块直接接收显式注入。
var globalApp *App

// Pool 返回当前进程的 *pgxpool.Pool。仅供过渡期使用。
func Pool() *pgxpool.Pool {
	if globalApp == nil {
		return nil
	}
	return globalApp.DB
}

// Config 返回当前进程的 *config.Config。仅供过渡期使用。
func Config() *config.Config {
	if globalApp == nil {
		return nil
	}
	return globalApp.Config
}

// Init 构造 App：打开数据库、加载配置、初始化缓存、session、权限服务。
func Init(cfg *config.Config) (*App, error) {
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

	app := &App{
		Config:      cfg,
		DB:          pool,
		SessionMgr:  sm,
		Server:      server.New(cfg),
		PermService: permService,
		Authz:       authzService,
		AppContext:  appCtx,
	}
	globalApp = app

	// 启动期引导
	if bcfg := LoadBootstrapConfig(); bcfg.Enabled {
		if err := RunBootstrap(ctx, pool, bcfg); err != nil {
			log.Printf("[bootstrap] failed: %v", err)
		}
	}

	return app, nil
}

// Shutdown 释放资源。
func Shutdown(app *App) {
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
	if globalApp == app {
		globalApp = nil
	}
}

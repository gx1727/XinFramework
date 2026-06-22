// Package boot 启动编排：构造业务模块依赖的 *appx.App（Config + DB），
// 以及 framework 内部使用的 HTTP server 与 AppContext。
//
// 资源分工：
//   - *appx.App：暴露给业务模块的最小依赖容器
//   - *server.XinServer / *plugin.AppContext：framework 内部运行时资源，
//     封装在 framework.Runtime 里，不传给业务模块
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

// Init 构造进程级资源并返回 (App, HTTP server, AppContext)。
//
//   - app 暴露给业务模块（Config + DB）；
//   - srv / appCtx 仅供 framework 内部使用，封装到 framework.Runtime 后传给
//     Serve / signal 处理等 framework 内部代码。
//
// Session / Authz 等共享资源由 appCtx 持有：Session 在 NewAppContext 时填充，
// Authz 在本函数末尾 SetAuthz 填充；framework 通过 rt.AppCtx.Session() /
// rt.AppCtx.Authz() 读取，不必再单独持有。
func Init(cfg *config.Config) (*appx.App, *server.XinServer, *plugin.AppContext, error) {
	logger.Init(cfg.Log.Dir, cfg.Log.Level)

	ctx := context.Background()
	pool, err := db.Init(ctx, &cfg.Database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("db init failed: %w", err)
	}

	dict.Init(pool)

	if err := cache.Init(&cfg.Redis); err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("cache init failed: %w", err)
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

	srv := server.New(cfg)

	app := &appx.App{
		Config: cfg,
		DB:     pool,
	}

	_ = permService // kept for potential future internal use; not exposed via App

	return app, srv, appCtx, nil
}

// Shutdown 释放基础设施资源（cache / DB pool / logger）。
//
// 模块级 Shutdown 由 framework.Serve 在收到信号后调用 shutdownModules 完成，
// 不在此处——boot 包不持有模块列表，模块生命周期归 framework 管。
func Shutdown(app *appx.App) {
	if app == nil {
		return
	}
	if err := cache.Close(); err != nil {
		log.Printf("cache close failed: %v", err)
	}
	if app.DB != nil {
		app.DB.Close()
	}
	logger.Close()
}
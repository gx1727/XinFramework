// Package boot 启动编排：构造业务模块依赖的 *appx.App（Config + DB），
// 以及 framework 内部使用的 HTTP server 与 AppContext。
//
// 资源分工：
//   - *appx.App：暴露给业务模块的最小依赖容器
//   - *server.Server / *plugin.AppContext：framework 内部运行时资源，
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
	"time"

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
// Authz 在本函数末尾 SetAuthz 填充；framework 通过 rt.AppCtx.Session() / rt.AppCtx.Authz() 读取，不必再单独持有。
func Init(cfg *config.Config) (*appx.App, *server.Server, *plugin.AppContext, error) {
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

	// 权限缓存装配：
	//   - Redis 可用 → 用 RedisPermissionCache（多实例共享缓存）
	//   - Redis 不可用 → fallback 到 MemoryPermissionCache（仅本进程有效）
	// 两种实现都满足 permission.PermissionCache 接口，PermissionService
	// 无需感知具体类型。
	var permCache permission.PermissionCache
	if cache.Get() != nil {
		permCache = permission.NewRedisPermissionCache()
	} else {
		memCache := permission.NewMemoryPermissionCache()
		if cfg.PermissionCache.PermTTLSeconds > 0 {
			memCache.SetPermTTL(time.Duration(cfg.PermissionCache.PermTTLSeconds) * time.Second)
		}
		if cfg.PermissionCache.DataScopeTTLSeconds > 0 {
			memCache.SetDataScopeTTL(time.Duration(cfg.PermissionCache.DataScopeTTLSeconds) * time.Second)
		}
		permCache = memCache
	}

	appCtx, err := plugin.NewAppContext(appx.MustNewPool(pool), cache.Get(), cfg, sm)
	if err != nil {
		pool.Close()
		return nil, nil, nil, fmt.Errorf("new app context: %w", err)
	}

	permService := service.NewPermissionService(
		permission.NewPermissionRepository(pool),
		permission.NewDataScopeRepository(pool),
		permCache,
		permission.NewSysRoleRepository(pool),
	)
	authzService := service.NewAuthorizationService(permService)
	// 编译期保证:*AuthorizationService 必须实现 authz.Authorization。
	// 如果以后改方法签名漏改接口,这里直接 build fail,而不是运行时 type assert 炸。
	var _ authz.Authorization = (*service.AuthorizationService)(nil)
	appCtx.SetAuthz(authzService)

	srv := server.New(cfg)

	// Phase 0024：用强类型 Pool 包装，构造期 fail-fast 消灭所有 nil-check
	// 散落到各 module 的 `if p := ctx.DB(); p != nil` 模式。
	app := appx.MustNewApp(cfg, appx.MustNewPool(pool))

	return app, srv, appCtx, nil
}

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

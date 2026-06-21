// Package framework 是显式 Build 阶段的总入口。
//
// 阶段设计（重构后）：
//  1. main.go 加载 config (config.Load)
//  2. main.go 构造 *appx.App (boot.Init)
//  3. main.go 显式调用每个模块的 Module(app) 拿到 []plugin.Module
//  4. main.go 调用 framework.Serve(cfg, app, modules) 启动服务
//
// 不再有 framework.Run(cfg) 这类把上面四步打包的便捷函数。
// 不再依赖 plugin.Apps() 的全局注册表 + side-effect import。
package framework

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/config"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
)

const (
	pidFile = "./xin.pid" // PID文件路径
	logFile = "./xin.log" // 日志文件路径
)

// Serve 是显式 Build 阶段之后的"运行"阶段。
//
// 调用方负责：
//   - 加载 config
//   - 调用 boot.Init(cfg) 拿到 *appx.App
//   - 显式构造各模块的 []plugin.Module 列表
//
// Serve 负责：
//   - 跑 migrate
//   - 对每个模块调 Init()，期间各模块向 app.AppContext 写自己的 repository
//   - 装全局中间件 + 调每个模块的 Register() 注册路由
//   - 启 HTTP server，阻塞到收到信号后优雅退出
func Serve(cfg *config.Config, app *appx.App, modules []plugin.Module) {
	if err := migrate.Run(app.DB, "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	enabled := enabledSet(cfg)

	// Init 阶段：每个模块把自己的依赖写进 AppContext
	for _, m := range modules {
		if !enabled[m.Name()] {
			log.Printf("module %s not enabled (skip init)", m.Name())
			continue
		}
		ctx, w := buildAppContextPair(app.AppContext)
		if err := m.Init(ctx, w); err != nil {
			log.Fatalf("module %s init failed: %v", m.Name(), err)
		}
		log.Printf("module %s initialized", m.Name())
	}

	// 配置全局中间件 + 路由
	setupRouter(app, modules)

	// 启动 HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("server starting on %s", addr)

	go func() {
		if err := app.Server.Start(addr); err != nil {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	if err := sdNotifyReady(); err != nil {
		log.Printf("sd_notify ready: %v", err)
	}

	waitForSignal(app.Server, app)
}

// Boot 公开包外可用的 boot 入口，封装 internal/core/boot.Init。
//
// main.go 没法直接 import internal 包，所以通过这个薄包装调用。
func Boot(cfg *config.Config) (*appx.App, error) {
	return boot.Init(cfg)
}

// enabledSet 把 cfg.Module 列表转成 set 方便 O(1) 查询。
func enabledSet(cfg *config.Config) map[string]bool {
	s := make(map[string]bool, len(cfg.Module))
	for _, name := range cfg.Module {
		s[name] = true
	}
	return s
}

// buildAppContextPair 构造 (Reader, Writer)。AppContext 同时实现这两个接口，
// 同一个指针传给两侧让模块能写自己负责的 slot 并读别的模块的 slot。
func buildAppContextPair(appCtx *plugin.AppContext) (plugin.Reader, plugin.Writer) {
	if appCtx == nil {
		return nil, nil
	}
	return appCtx, appCtx
}

func setupRouter(app *appx.App, modules []plugin.Module) {
	srv := app.Server
	cfg := app.Config

	// 注册全局中间件（按执行顺序）
	srv.Engine.Use(middleware.Recovery())      // 1. 异常恢复，最先执行以捕获所有下游 panic
	srv.Engine.Use(middleware.RequestID())     // 2. 请求ID，尽早标记每次请求
	srv.Engine.Use(middleware.CORS(&cfg.CORS)) // 3. CORS 预检请求处理
	srv.Engine.Use(middleware.ClientIP())      // 4. 客户端 IP 注入 ctx（供 audit 使用）
	srv.Engine.Use(middleware.Logger())        // 5. 日志（依赖 RequestID）

	// 注册所有模块的路由
	registerModules(srv.Engine, cfg, app, modules)
}

// registerModules 注册已启用模块的路由（所有模块统一处理，无内置/外部之分）。
//
// 路由空间（重构后）：
//   - public     → /api/v1/*             （OptionalAuth，公开；需隔离的子资源挂 /public/<x>）
//   - tenant     → /api/v1/*             （Auth + RequireTenantContext，业务域；模块直接挂资源路径，无 /t 前缀）
//   - protected  → /api/v1/platform/*    （Auth + RequirePlatformRole，平台域）
//
// 三组 RouterGroup 都通过 plugin.Module.Register(ctx, public, tenant, protected)
// 传给业务模块，由模块自行选择挂在哪一组。
func registerModules(r *gin.Engine, cfg *config.Config, app *appx.App, modules []plugin.Module) {
	v1 := r.Group("/api/v1")

	// public：可选登录，公开读
	public := v1.Group("")
	public.Use(middleware.OptionalAuth(&cfg.JWT, app.SessionMgr, app.Authz, app.DB))

	// tenant：必须登录 + 必须携带有效 tenant_id > 0（业务域）
	// 挂载点为 ""——模块直接挂资源路径，例如 tenant.Group("/users") → /api/v1/users
	tenant := v1.Group("")
	tenant.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.Authz, app.DB))
	tenant.Use(pkgmiddleware.RequireTenantContext())

	// protected：必须登录（语义上是 platform 域）
	// 平台模块（platform_tenant / platform_menu / config platform 域 / dict platform 域）
	// 自己在内部追加 RequirePlatformRole("super_admin")
	protected := v1.Group("/platform")
	protected.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.Authz, app.DB))

	enabled := enabledSet(cfg)
	// Same Reader that Init used. By the time we reach Register,
	// every module has finished its Init phase, so ctx.Reader exposes
	// all repositories that were populated during Init.
	var ctx plugin.Reader = app.AppContext

	for _, m := range modules {
		if !enabled[m.Name()] {
			log.Printf("module %s not enabled (skip register)", m.Name())
			continue
		}
		// 三组 RouterGroup（public / tenant / protected）
		// - public:    公开路由（/auth/*、/health、/public/configs 等）
		// - tenant:    业务域路由（/users、/menus、/configs 等；无 /t 前缀）
		// - protected: 平台域路由（/platform/configs、/platform/dicts、/platform/tenants、/platform/menus）
		m.Register(ctx, public, tenant, protected)
	}
}

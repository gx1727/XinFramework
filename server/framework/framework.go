// Package framework 是显式 Build 阶段的总入口。
//
// 阶段设计（重构后）：
//  1. main.go 加载 config (config.Load)
//  2. main.go 构造 (*appx.App, *Runtime) (framework.Boot → boot.Init)
//  3. main.go 显式调用每个模块的 Module(app) 拿到 []plugin.Module
//  4. main.go 调用 framework.Serve(cfg, app, rt, modules) 启动服务
//
// 不再有 framework.Run(cfg) 这类把上面四步打包的便捷函数。
// 不再依赖任何全局注册表——modules 是显式参数，从 main.go 一路传到 Serve / Shutdown。
package framework

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
//   - 调用 framework.Boot(cfg) 拿到 (*appx.App, *Runtime)
//   - 显式构造各模块的 []plugin.Module 列表
//
// Serve 负责：
//   - 跑 migrate
//   - 对每个模块调 Init()，期间各模块向 rt.AppCtx 写自己的 repository
//   - 装全局中间件 + 调每个模块的 Register() 注册路由
//   - 启 HTTP server，阻塞到收到信号后优雅退出
//   - 信号触发后调各模块 Shutdown()，再释放基础设施
func Serve(cfg *config.Config, app *appx.App, rt *Runtime, modules []plugin.Module) {
	if err := migrate.Run(app.DB.Raw(), "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	enabled := enabledSet(cfg)

	// Init 阶段：每个模块把自己的依赖写进 rt.AppCtx
	for _, m := range modules {
		if !enabled[m.Name()] {
			log.Printf("module %s not enabled (skip init)", m.Name())
			continue
		}
		ctx, w := buildAppContextPair(rt.AppCtx)
		if err := m.Init(ctx, w); err != nil {
			log.Fatalf("module %s init failed: %v", m.Name(), err)
		}
		log.Printf("module %s initialized", m.Name())
	}

	// 配置全局中间件 + 路由
	setupRouter(cfg, app.DB.Raw(), rt, modules)

	// 启动 HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("server starting on %s", addr)

	go func() {
		if err := rt.Server.Start(addr); err != nil {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	if err := sdNotifyReady(); err != nil {
		log.Printf("sd_notify ready: %v", err)
	}

	waitForSignal(rt, app, modules)
}

// Boot 公开包外可用的 boot 入口，封装 internal/core/boot.Init。
//
// main.go 没法直接 import internal 包，所以通过这个薄包装调用。
// 返回 (*appx.App, *Runtime, error)：
//   - app 传给每个模块的 Module(app) 构造函数；
//   - rt 供 framework 内部（HTTP server / 模块装配 / 信号处理）使用。
func Boot(cfg *config.Config) (*appx.App, *Runtime, error) {
	app, srv, appCtx, err := boot.Init(cfg)
	if err != nil {
		return nil, nil, err
	}
	return app, &Runtime{Server: srv, AppCtx: appCtx}, nil
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

func setupRouter(cfg *config.Config, db *pgxpool.Pool, rt *Runtime, modules []plugin.Module) {
	r := rt.Server.Engine

	// 注册全局中间件（按执行顺序）
	r.Use(middleware.Recovery())      // 1. 异常恢复，最先执行以捕获所有下游 panic
	r.Use(middleware.RequestID())     // 2. 请求ID，尽早标记每次请求
	r.Use(middleware.CORS(&cfg.CORS)) // 3. CORS 预检请求处理
	r.Use(middleware.ClientIP())      // 4. 客户端 IP 注入 ctx（供 audit 使用）
	r.Use(middleware.Logger())        // 5. 日志（依赖 RequestID）

	// 注册所有模块的路由
	registerModules(r, cfg, db, rt.AppCtx, modules)
}

// registerModules 注册已启用模块的路由
//
// 路由空间（重构后）：
//   - public     → /api/v1/*             （OptionalAuth，公开；需隔离的子资源挂 /public/<x>）
//   - tenant     → /api/v1/*             （Auth + RequireTenantContext，业务域；模块直接挂资源路径）
//   - protected  → /api/v1/sys/*         （Auth + RequireSysRole，sys 域）
//
// 三组 RouterGroup 打包成 plugin.RouterSlots 传给业务模块。
// 业务模块用 slots.MustGet(plugin.SlotPublic | SlotTenant | SlotProtected) 取。
func registerModules(r *gin.Engine, cfg *config.Config, db *pgxpool.Pool, appCtx *plugin.AppContext, modules []plugin.Module) {
	v1 := r.Group("/api/v1")

	// Session / Authz 由 AppCtx 持有——boot.Init 已分别在 NewAppContext 和
	// SetAuthz 阶段填充，Register 阶段读取必非 nil。
	sm := appCtx.Session()
	authzSvc := appCtx.Authz()

	// public：可选登录，公开读
	public := v1.Group("")
	public.Use(middleware.OptionalAuth(&cfg.JWT, sm, authzSvc, db))

	// tenant：必须登录 + 必须携带有效 tenant_id > 0（业务域）
	// 挂载点为 ""——模块直接挂资源路径，例如 tenant.Group("/users") → /api/v1/users
	tenant := v1.Group("")
	tenant.Use(middleware.Auth(&cfg.JWT, sm, authzSvc, db))
	tenant.Use(pkgmiddleware.RequireTenantContext())

	// protected：必须登录（语义上是 sys 域）
	// 平台模块（sys_tenant / sys_menu / config sys 域 / dict sys 域）
	// 自己在内部追加 RequireSysRole("super_admin")
	protected := v1.Group("/sys")
	protected.Use(middleware.Auth(&cfg.JWT, sm, authzSvc, db))

	enabled := enabledSet(cfg)
	// Same Reader that Init used. By the time we reach Register,
	// every module has finished its Init phase, so ctx.Reader exposes
	// all repositories that were populated during Init.
	var ctx plugin.Reader = appCtx

	// 打包为 slots。后续要新增第 4 类路由（如 /api/v2）只需在这里往 map 里加一项。
	slots := plugin.RouterSlots{
		plugin.SlotPublic: {
			Name:        plugin.SlotPublic,
			Group:       public,
			Description: "公开接口（OptionalAuth）",
		},
		plugin.SlotTenant: {
			Name:        plugin.SlotTenant,
			Group:       tenant,
			Description: "租户域业务（Auth + RequireTenantContext）",
		},
		plugin.SlotProtected: {
			Name:        plugin.SlotProtected,
			Group:       protected,
			Description: "sys 域管理（Auth）",
		},
	}

	for _, m := range modules {
		if !enabled[m.Name()] {
			log.Printf("module %s not enabled (skip register)", m.Name())
			continue
		}
		m.Register(ctx, slots)
	}
}

// shutdownModules 在收到关闭信号时按顺序调用每个模块的 Shutdown。
//
// 模块 Shutdown 接收的 reader 为 nil——目前没有模块需要它；将来若有
// "读别人 slot 来清理自己"的场景，再扩成传 rt.AppCtx。
func shutdownModules(modules []plugin.Module) {
	var reader plugin.Reader
	for _, m := range modules {
		if err := m.Shutdown(reader); err != nil {
			log.Printf("module %s shutdown failed: %v", m.Name(), err)
		}
	}
}

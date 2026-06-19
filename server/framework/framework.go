package framework

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
)

const (
	pidFile = "./xin.pid" // PID文件路径
	logFile = "./xin.log" // 日志文件路径
)

// RegisterModule registers a module with the framework. This is now a
// thin wrapper around plugin.Register — there is no longer a separate
// "builtin" list. All modules, regardless of whether they ship with
// the framework or live under apps/, register through this single
// entry point.
func RegisterModule(m plugin.Module) {
	plugin.Register(m)
}

// Run 框架入口函数，根据命令行参数执行相应操作
func Run(cfg *config.Config) {
	if len(os.Args) < 2 {
		runServer(cfg)
		return
	}

	switch os.Args[1] {
	case "start":
		cmdStart()
	case "stop":
		cmdStop()
	case "restart":
		cmdRestart()
	case "reload":
		cmdReload()
	case "status":
		cmdStatus()
	case "hot-restart":
		cmdHotRestart()
	case "run":
		runServer(cfg)
	case "help", "-h", "--help":
		printUsage()
	default:
		printUsage()
	}
}

func runServer(cfg *config.Config) {
	app, err := boot.Init(cfg)
	if err != nil {
		log.Fatalf("boot init failed: %v", err)
	}

	// 初始化所有模块（统一列表，不再区分 builtin / external）
	if err := initModules(app); err != nil {
		log.Fatalf("module init failed: %v", err)
	}

	// 执行数据迁移（pool 由 boot 持有）
	if err := migrate.Run(app.DB, "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// 配置路由和中间件
	setupRouter(app)

	// 构建服务器地址
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("server starting on %s", addr)

	// 在后台启动HTTP服务器
	go func() {
		if err := app.Server.Start(addr); err != nil {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	// 发送systemd就绪通知（如果支持）
	if err := sdNotifyReady(); err != nil {
		log.Printf("sd_notify ready: %v", err)
	}

	// 等待系统信号（用于优雅关闭）
	waitForSignal(app.Server, app)
}

// initModules iterates through every registered module.
//
// 之前内置模块在 builtinMap 中查找、外部 app 走 plugin.Apps()，现在
// 两者合二为一。cfg.Module 控制加载哪些模块；未在配置中启用的模块
// 也会注册但跳过 Init / Register，避免误启用。
func initModules(app *boot.App) error {
	cfg := app.Config
	enabled := make(map[string]bool, len(cfg.Module))
	for _, name := range cfg.Module {
		enabled[name] = true
	}

	for _, m := range plugin.Apps() {
		if !enabled[m.Name()] {
			log.Printf("module %s registered but not enabled (skip)", m.Name())
			continue
		}
		// Build the (Reader, Writer) pair from the shared AppContext
		// that boot.Init constructed. Every module sees the same
		// instance so writes one module makes are visible to readers
		// of any other module.
		ctx, w := buildAppContextForModule(app.AppContext)
		if err := m.Init(ctx, w); err != nil {
			return fmt.Errorf("module %s init failed: %w", m.Name(), err)
		}
		log.Printf("module %s initialized", m.Name())
	}
	return nil
}

// buildAppContextForModule constructs (Reader, Writer) for a module.
// AppContext is a concrete struct that satisfies both Reader and
// Writer; passing the same pointer to both sides lets modules write
// to the slot they own while reading slots contributed by others.
func buildAppContextForModule(appCtx *plugin.AppContext) (plugin.Reader, plugin.Writer) {
	if appCtx == nil {
		return nil, nil
	}
	return appCtx, appCtx
}

func setupRouter(app *boot.App) {
	srv := app.Server
	cfg := app.Config

	// 注册全局中间件（按执行顺序）
	srv.Engine.Use(middleware.Recovery())      // 1. 异常恢复，最先执行以捕获所有下游 panic
	srv.Engine.Use(middleware.RequestID())     // 2. 请求ID，尽早标记每次请求
	srv.Engine.Use(middleware.CORS(&cfg.CORS)) // 3. CORS 预检请求处理
	srv.Engine.Use(middleware.ClientIP())      // 4. 客户端 IP 注入 ctx（供 audit 使用）
	srv.Engine.Use(middleware.Logger())        // 5. 日志（依赖 RequestID）

	// 注册所有模块的路由
	registerModules(srv.Engine, cfg, app)
}

// registerModules 注册已启用模块的路由（所有模块统一处理，无内置/外部之分）。
func registerModules(r *gin.Engine, cfg *config.Config, app *boot.App) {
	v1 := r.Group("/api/v1")
	public := v1.Group("")
	public.Use(middleware.OptionalAuth(&cfg.JWT, app.SessionMgr, app.Authz, app.DB))

	protected := v1.Group("")
	protected.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.Authz, app.DB))

	enabled := make(map[string]bool, len(cfg.Module))
	for _, name := range cfg.Module {
		enabled[name] = true
	}

	// Same Reader that Init used. By the time we reach Register,
	// every module has finished its Init phase, so ctx.Reader exposes
	// all repositories that were populated during Init.
	var ctx plugin.Reader = app.AppContext

	for _, m := range plugin.Apps() {
		if !enabled[m.Name()] {
			continue
		}
		m.Register(ctx, public, protected)
	}
}
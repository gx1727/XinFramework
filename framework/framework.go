package framework

import (
	"fmt"
	"log"
	"os"

	v1 "gx1727.com/xin/framework/api/v1"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/internal/module/tenant"
	"gx1727.com/xin/framework/internal/module/user"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
)

const (
	pidFile = "./xin.pid" // PID文件路径
	logFile = "./xin.log" // 日志文件路径
)

// RegisterModule 注册插件模块到框架
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

	// 初始化所有插件模块
	initModules()

	// 执行数据迁移
	runMigrations()

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

// runFrameworkMigrations 执行框架核心数据库迁移
func runMigrations() {
	if err := migrate.Run("migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
}

func initModules() {
	for _, m := range plugin.All() {
		if err := m.Init(); err != nil {
			log.Fatalf("module %s init failed: %v", m.Name(), err)
		}
		log.Printf("module %s initialized", m.Name())
	}
}

func setupRouter(app *boot.App) {
	srv := app.Server
	cfg := app.Config

	// 注册全局中间件
	srv.Engine.Use(middleware.CORS(&cfg.CORS))       // CORS 中间件
	srv.Engine.Use(middleware.RequestID())           // 请求ID中间件
	srv.Engine.Use(middleware.Logger())              // 日志中间件
	srv.Engine.Use(middleware.Recovery())            // 异常恢复中间件
	srv.Engine.Use(middleware.Tenant(cfg.Saas.Mode)) // 租户中间件

	// 构建内置模块处理器
	handlers := buildBuiltinHandlers(app)

	v1.RegisterRoutes(srv.Engine, cfg, app.SessionMgr, v1.Dependencies{
		UserHandler:   handlers["user"].(*user.Handler),
		TenantHandler: handlers["tenant"].(*tenant.Handler),
	})
}

// builtinHandlerBuilder creates handlers for built-in modules given app context
type builtinHandlerBuilder func(*boot.App) interface{}

var builtinHandlers = map[string]builtinHandlerBuilder{
	"user": func(app *boot.App) interface{} {
		repos := user.Repositories{
			Account: app.Repository.Account(),
			Tenant:  app.Repository.Tenant(),
			Role:    app.Repository.Role(),
			User:    app.Repository.User(),
		}
		deps := user.DefaultDependencies(app.Config, app.DB, repos)
		return user.NewHandler(user.NewService(deps))
	},
	"tenant": func(app *boot.App) interface{} {
		return tenant.NewHandler(tenant.NewService(app.Repository.Tenant()))
	},
}

func buildBuiltinHandlers(app *boot.App) map[string]interface{} {
	result := make(map[string]interface{})
	for name, builder := range builtinHandlers {
		result[name] = builder(app)
	}
	return result
}

// RegisterBuiltinHandler registers a handler builder for a built-in module
func RegisterBuiltinHandler(name string, builder builtinHandlerBuilder) {
	builtinHandlers[name] = builder
}

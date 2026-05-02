package framework

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/core/middleware"
	assetModule "gx1727.com/xin/framework/internal/module/asset"
	authModule "gx1727.com/xin/framework/internal/module/auth"
	dictModule "gx1727.com/xin/framework/internal/module/dict"
	menuModule "gx1727.com/xin/framework/internal/module/menu"
	orgModule "gx1727.com/xin/framework/internal/module/organization"
	permModule "gx1727.com/xin/framework/internal/module/permission"
	resourceModule "gx1727.com/xin/framework/internal/module/resource"
	roleModule "gx1727.com/xin/framework/internal/module/role"
	systemModule "gx1727.com/xin/framework/internal/module/system"
	tenantModule "gx1727.com/xin/framework/internal/module/tenant"
	userModule "gx1727.com/xin/framework/internal/module/user"
	weixinModule "gx1727.com/xin/framework/internal/module/weixin"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
)

const (
	pidFile = "./xin.pid" // PID文件路径
	logFile = "./xin.log" // 日志文件路径
)

var builtinMap = map[string]plugin.Module{
	"asset":        assetModule.Module(),
	"auth":         authModule.Module(),
	"tenant":       tenantModule.Module(),
	"user":         userModule.Module(),
	"menu":         menuModule.Module(),
	"dict":         dictModule.Module(),
	"role":         roleModule.Module(),
	"resource":     resourceModule.Module(),
	"organization": orgModule.Module(),
	"permission":   permModule.Module(),
	"system":       systemModule.Module(),
	"weixin":       weixinModule.Module(),
}

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

	// 初始化所有模块（内置 + 外部）
	initModules(cfg)

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

func initModules(cfg *config.Config) {
	for _, name := range cfg.Module {
		if m, ok := builtinMap[name]; ok {
			if err := m.Init(); err != nil {
				log.Fatalf("builtin module %s init failed: %v", name, err)
			}
			log.Printf("builtin module %s initialized", name)
		} else {
			log.Fatalf("configured builtin module %s not found", name)
		}
	}

	// 初始化外部插件模块
	for _, m := range plugin.Apps() {
		if err := m.Init(); err != nil {
			log.Fatalf("module %s init failed: %v", m.Name(), err)
		}
		log.Printf("module %s initialized", m.Name())
	}
}

func runMigrations() {
	if err := migrate.Run("migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
}

func setupRouter(app *boot.App) {
	srv := app.Server
	cfg := app.Config

	// 注册全局中间件（按执行顺序）
	srv.Engine.Use(middleware.Recovery())            // 1. 异常恢复，最先执行以捕获所有下游 panic
	srv.Engine.Use(middleware.RequestID())           // 2. 请求ID，尽早标记每次请求
	srv.Engine.Use(middleware.CORS(&cfg.CORS))       // 3. CORS 预检请求处理
	srv.Engine.Use(middleware.Logger())              // 4. 日志（依赖 RequestID）
	srv.Engine.Use(middleware.Tenant(cfg.Saas.Mode)) // 5. 租户上下文

	// 注册内置模块和外部插件
	registerModules(srv.Engine, cfg, app)
}

// registerModules 注册内置模块和外部插件的路由
func registerModules(r *gin.Engine, cfg *config.Config, app *boot.App) {
	v1 := r.Group("/api/v1")
	public := v1.Group("")
	public.Use(middleware.OptionalAuth(&cfg.JWT, app.SessionMgr, app.PermService))

	protected := v1.Group("")
	protected.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.PermService))

	// 注册内置模块路由
	for _, name := range cfg.Module {
		if m, ok := builtinMap[name]; ok {
			m.Register(public, protected)
		}
	}

	// 注册外部插件路由
	for _, m := range plugin.Apps() {
		m.Register(public, protected)
	}
}

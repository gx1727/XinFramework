package framework

import (
	"fmt"
	"log"
	"os"

	v1 "gx1727.com/xin/framework/api/v1"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/internal/core/server"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
)

const (
	pidFile = "./xin.pid"
	logFile = "./xin.log"
)

func RegisterModule(m plugin.Module) {
	plugin.Register(m)
}

func Run() {
	if len(os.Args) < 2 {
		runServer()
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
		runServer()
	case "help", "-h", "--help":
		printUsage()
	default:
		printUsage()
	}
}

func runServer() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	srv, err := boot.Init(cfg)
	if err != nil {
		log.Fatalf("boot init failed: %v", err)
	}

	initModules()
	runFrameworkMigrations()
	migrateModules()

	setupRouter(srv, cfg)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("server starting on %s", addr)

	go func() {
		if err := srv.Start(addr); err != nil {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	if err := sdNotifyReady(); err != nil {
		log.Printf("sd_notify ready: %v", err)
	}

	waitForSignal(srv)
}

func runFrameworkMigrations() {
	if err := migrate.Run("migrations"); err != nil {
		log.Fatalf("framework migrations failed: %v", err)
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

func migrateModules() {
	for _, m := range plugin.All() {
		if err := m.Migrate(); err != nil {
			log.Fatalf("module %s migrate failed: %v", m.Name(), err)
		}
	}
}

func setupRouter(srv *server.XinServer, cfg *config.Config) {
	srv.Engine.Use(middleware.RequestID())
	srv.Engine.Use(middleware.Logger())
	srv.Engine.Use(middleware.Recovery())
	srv.Engine.Use(middleware.Tenant(cfg.Saas.Mode))

	v1.RegisterRoutes(srv.Engine, cfg)
}

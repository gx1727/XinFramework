package main

import (
	"fmt"
	"gx1727.com/xin/api/v1"
	"gx1727.com/xin/internal/core/boot"
	"gx1727.com/xin/internal/core/middleware"
	"gx1727.com/xin/internal/core/server"
	"gx1727.com/xin/pkg/config"
	"gx1727.com/xin/pkg/resp"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	pidFile = "./xin.pid"
	logFile = "./xin.log"
)

func main() {
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

func setupRouter(srv *server.XinServer, cfg *config.Config) {
	srv.Engine.Use(middleware.RequestID())
	srv.Engine.Use(middleware.Logger())
	srv.Engine.Use(middleware.Recovery())
	srv.Engine.Use(middleware.Tenant(cfg.Saas.Mode))

	v1.RegisterRoutes(srv.Engine, cfg)

	auth := srv.Engine.Group("/api/v1")
	auth.Use(middleware.Auth(&cfg.JWT))
	{
		auth.GET("/users", func(c *gin.Context) {
			resp.Error(c, 1001, "not implemented")
		})
		auth.POST("/users", func(c *gin.Context) {
			resp.Error(c, 1001, "not implemented")
		})
		auth.PUT("/users/:id", func(c *gin.Context) {
			resp.Error(c, 1001, "not implemented")
		})
		auth.DELETE("/users/:id", func(c *gin.Context) {
			resp.Error(c, 1001, "not implemented")
		})
	}
}

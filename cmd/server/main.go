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
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
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

	waitForSignal()
}

func setupRouter(srv *server.XinServer, cfg *config.Config) {
	srv.Engine.Use(middleware.Logger())
	srv.Engine.Use(middleware.Recovery())
	srv.Engine.Use(middleware.Tenant(cfg.Saas.Mode))

	v1.RegisterRoutes(srv.Engine)

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

func waitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	sig := <-sigCh
	log.Printf("Received signal: %v", sig)

	boot.GetServer().Shutdown(30 * time.Second)

	log.Printf("Server exited gracefully")
	os.Exit(0)
}

func sdNotifyReady() error {
	if os.Getenv("NOTIFY_SOCKET") == "" {
		return nil
	}

	socketPath := os.Getenv("NOTIFY_SOCKET")
	conn, err := net.Dial("unixgram", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("READY=1"))
	return err
}

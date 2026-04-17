package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xin-framework/xin/configs"
	"github.com/xin-framework/xin/internal/core/boot"
	"github.com/xin-framework/xin/internal/core/middleware"
	"github.com/xin-framework/xin/internal/core/server"
)

func main() {
	cfg, err := configs.Load("configs/config.yaml")
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
	if err := srv.Start(addr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}

func setupRouter(srv *server.XinServer, cfg *configs.Config) {
	srv.Engine.Use(middleware.Logger())
	srv.Engine.Use(middleware.Recovery())
	srv.Engine.Use(middleware.Tenant())

	auth := srv.Engine.Group("/api/v1")
	auth.Use(middleware.Auth(&cfg.JWT))
	{
		auth.GET("/users", func(c *gin.Context) {})
		auth.POST("/users", func(c *gin.Context) {})
		auth.PUT("/users/:id", func(c *gin.Context) {})
		auth.DELETE("/users/:id", func(c *gin.Context) {})
	}
}

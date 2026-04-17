package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gx1727.com/xin-framework/configs"
	"gx1727.com/xin-framework/internal/core/boot"
	"gx1727.com/xin-framework/internal/core/middleware"
	"gx1727.com/xin-framework/internal/core/server"
	"log"
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

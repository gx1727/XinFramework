package main

import (
	"fmt"
	"gx1727.com/xin/api/v1"
	"gx1727.com/xin/internal/core/boot"
	"gx1727.com/xin/internal/core/middleware"
	"gx1727.com/xin/internal/core/server"
	"gx1727.com/xin/pkg/config"
	"gx1727.com/xin/pkg/resp"

	"github.com/gin-gonic/gin"
	"log"
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
	if err := srv.Start(addr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
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
			resp.Error(c, 501, "not implemented")
		})
		auth.POST("/users", func(c *gin.Context) {
			resp.Error(c, 501, "not implemented")
		})
		auth.PUT("/users/:id", func(c *gin.Context) {
			resp.Error(c, 501, "not implemented")
		})
		auth.DELETE("/users/:id", func(c *gin.Context) {
			resp.Error(c, 501, "not implemented")
		})
	}
}

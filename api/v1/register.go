package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/internal/core/middleware"
	"gx1727.com/xin/internal/module/auth"
	"gx1727.com/xin/internal/module/cms"
	"gx1727.com/xin/internal/module/system"
	"gx1727.com/xin/internal/module/weixin"
	"gx1727.com/xin/pkg/config"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	v1 := r.Group("/api/v1")

	public := v1.Group("")
	protected := v1.Group("")
	protected.Use(middleware.Auth(&cfg.JWT))

	auth.RegisterV1(public, protected)

	if cfg.DomainEnabled("system") {
		system.RegisterV1(public, protected)
	}
	if cfg.DomainEnabled("cms") {
		cms.RegisterV1(protected)
	}
	if cfg.DomainEnabled("weixin") {
		weixin.RegisterV1(protected)
	}
}

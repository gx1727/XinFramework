package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/internal/module/auth"
	"gx1727.com/xin/framework/internal/module/system"
	"gx1727.com/xin/framework/internal/module/weixin"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
)

var builtinModules = []plugin.Module{
	auth.Module(),
	system.Module(),
	weixin.Module(),
}

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	v1 := r.Group("/api/v1")

	public := v1.Group("")
	protected := v1.Group("")
	protected.Use(middleware.Auth(&cfg.JWT))

	for _, m := range builtinModules {
		if cfg.DomainEnabled(m.Name()) {
			m.Register(public, protected)
		}
	}

	for _, m := range plugin.All() {
		if cfg.DomainEnabled(m.Name()) {
			m.Register(public, protected)
		}
	}
}

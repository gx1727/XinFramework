package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/internal/module/system"
	"gx1727.com/xin/framework/internal/module/user"
	"gx1727.com/xin/framework/internal/module/weixin"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
)

type Dependencies struct {
	UserHandler *user.Handler
}

func builtinModules(deps Dependencies) []plugin.Module {
	return []plugin.Module{
		user.Module(deps.UserHandler),
		system.Module(),
		weixin.Module(),
	}
}

func RegisterRoutes(r *gin.Engine, cfg *config.Config, deps Dependencies) {
	v1 := r.Group("/api/v1")

	public := v1.Group("")
	protected := v1.Group("")
	protected.Use(middleware.Auth(&cfg.JWT))

	for _, m := range builtinModules(deps) {
		if cfg.ModuleEnabled(m.Name()) {
			m.Register(public, protected)
		}
	}

	for _, m := range plugin.All() {
		if cfg.ModuleEnabled(m.Name()) {
			m.Register(public, protected)
		}
	}
}

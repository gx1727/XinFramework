package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/internal/module/asset"
	"gx1727.com/xin/framework/internal/module/auth"
	"gx1727.com/xin/framework/internal/module/dict"
	"gx1727.com/xin/framework/internal/module/menu"
	"gx1727.com/xin/framework/internal/module/organization"
	"gx1727.com/xin/framework/internal/module/permission"
	"gx1727.com/xin/framework/internal/module/resource"
	"gx1727.com/xin/framework/internal/module/role"
	"gx1727.com/xin/framework/internal/module/system"
	"gx1727.com/xin/framework/internal/module/tenant"
	"gx1727.com/xin/framework/internal/module/user"
	"gx1727.com/xin/framework/internal/module/weixin"
	"gx1727.com/xin/framework/internal/service"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/session"
)

type Dependencies struct {
	AssetHandler        *asset.FileHandler
	AuthHandler         *auth.Handler
	TenantHandler       *tenant.Handler
	UserHandler         *user.Handler
	MenuHandler         *menu.Handler
	DictHandler         *dict.Handler
	RoleHandler         *role.Handler
	ResourceHandler     *resource.Handler
	OrganizationHandler *organization.Handler
	PermHandler         *permission.Handler
	PermService         *service.PermissionService
	WeixinHandler       *weixin.Handler
}

func builtinModules(deps Dependencies) []plugin.Module {
	return []plugin.Module{
		asset.Module(deps.AssetHandler),
		auth.Module(deps.AuthHandler),
		tenant.Module(deps.TenantHandler),
		user.Module(deps.UserHandler),
		menu.Module(deps.MenuHandler),
		dict.Module(deps.DictHandler),
		role.Module(deps.RoleHandler),
		resource.Module(deps.ResourceHandler),
		organization.Module(deps.OrganizationHandler),
		permission.Module(deps.PermHandler),
		system.Module(),
		weixin.Module(deps.WeixinHandler),
	}
}

func RegisterRoutes(r *gin.Engine, cfg *config.Config, sm session.SessionManager, deps Dependencies) {
	v1 := r.Group("/api/v1")

	public := v1.Group("")
	protected := v1.Group("")
	protected.Use(middleware.Auth(&cfg.JWT, sm, deps.PermService))

	for _, m := range builtinModules(deps) {
		if cfg.ModuleEnabled(m.Name()) {
			m.Register(public, protected)
		}
	}

	for _, m := range plugin.All() {
		m.Register(public, protected)
	}
}

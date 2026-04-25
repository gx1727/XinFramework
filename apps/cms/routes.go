package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	protected.GET("/cms/ping", h.Ping)
	protected.GET("/cms/user", h.GetUserByID)
	protected.GET("/cms/users", h.ListTenantUsers)
	protected.GET("/cms/tenant", h.GetTenant)
	protected.GET("/cms/users/search", h.SearchUsers)
	protected.GET("/cms/me", h.GetCurrentUser)
}

func Module(h *Handler) plugin.Module {
	return plugin.NewModuleWithOpts("cms",
		func(public *gin.RouterGroup, protected *gin.RouterGroup) {
			Register(public, protected, h)
		},
		plugin.WithInit(InitConfig),
	)
}

package auth

import (
	"gx1727.com/xin/framework/pkg/config"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/module/tenant"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Module() plugin.Module {
	return plugin.NewModule("auth", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		repos := Repositories{
			Account: NewAccountRepository(db.Get()),
			Tenant:  tenant.NewTenantRepository(db.Get()),
		}
		deps := DefaultDependencies(config.Get(), db.Get(), repos)
		h := NewHandler(NewService(deps))
		Register(public, protected, h)
	})
}

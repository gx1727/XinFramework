package role

import (
	"github.com/gin-gonic/gin"
	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())

	// Phase 3: register with framework/pkg/rbac so cross-module
	// consumers (framework/internal) can resolve role data without
	// importing apps/.
	pkgrbac.RegisterRoleRepository(func() pkgrbac.RoleRepository {
		return NewRoleRepository(db.Get())
	})
}

// Module 返回 role 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("role", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(
			NewRoleRepository(db.Get()),
			permission.NewDataScopeRepository(db.Get()),
			NewRoleMenuRepository(db.Get()),
		))
		Register(protected, h)
	})
}

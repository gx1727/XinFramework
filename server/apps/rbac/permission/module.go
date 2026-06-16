package permission

import (
	"github.com/gin-gonic/gin"
	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())

	// Phase 3: register with framework/pkg/rbac so Auth middleware
	// can resolve effective permissions via the public hook.
	pkgrbac.RegisterPermissionRepository(func() pkgrbac.RoleResourceRepository {
		return NewRoleResourceRepository(db.Get())
	})
}

// Module 返回 permission 模块的完整定义
// 管理角色-资源（按钮/API）权限，通过 role_resources 表
// 菜单权限管理已迁移到 role 模块
func Module() plugin.Module {
	return plugin.NewModule("permission", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		roleResourceRepo := NewRoleResourceRepository(db.Get())
		h := NewHandler(NewService(db.Get(), roleResourceRepo))
		Register(protected, h)
	})
}

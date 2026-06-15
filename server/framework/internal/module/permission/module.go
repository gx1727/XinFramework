package permission

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
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

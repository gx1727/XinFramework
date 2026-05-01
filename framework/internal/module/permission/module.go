package permission

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/module/menu"
	"gx1727.com/xin/framework/internal/module/resource"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 permission 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("permission", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		permRepo := permission.NewRolePermissionRepository(db.Get())
		h := NewHandler(NewService(db.Get(), permRepo, menu.NewMenuRepository(db.Get()), resource.NewResourceRepository(db.Get())))
		Register(protected, h)
	})
}

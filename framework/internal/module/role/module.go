package role

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 role 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("role", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(NewRoleRepository(db.Get()), permission.NewDataScopeRepository(db.Get())))
		Register(protected, h)
	})
}

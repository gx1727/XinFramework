package resource

import (
	"gx1727.com/xin/framework/internal/module/menu"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 resource 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("resource", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(NewResourceRepository(db.Get()), menu.NewMenuRepository(db.Get())))
		Register(protected, h)
	})
}

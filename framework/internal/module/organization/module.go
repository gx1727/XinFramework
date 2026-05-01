package organization

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 organization 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("organization", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(NewOrganizationRepository(db.Get())))
		Register(protected, h)
	})
}

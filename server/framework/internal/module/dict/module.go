package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 dict 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("dict", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewDictRepository(db.Get()))
		Register(protected, h)
	})
}

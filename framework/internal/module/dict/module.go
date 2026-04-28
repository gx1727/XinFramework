package dict

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/repository"
	"gx1727.com/xin/framework/pkg/plugin"
)

// Module 返回 dict 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("dict", func(public, protected *gin.RouterGroup) {
		h := NewHandler(repository.NewDictRepository(app.DB))
		Register(protected, h)
	})
}

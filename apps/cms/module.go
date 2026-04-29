package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/module/cms/internal/handler"
	"gx1727.com/xin/module/cms/internal/service"
)

var (
	cmsService *service.Service
	cmsHandler *handler.Handler
)

func init() {
	// 模块加载时初始化 Service 和 Handler
	cmsService = service.NewService()
	cmsHandler = handler.NewHandler(cmsService)
}

// Module 返回 CMS 插件模块
func Module() plugin.Module {
	return plugin.NewModule("cms", func(public, protected *gin.RouterGroup) {
		Register(cmsHandler, public, protected)
	})
}

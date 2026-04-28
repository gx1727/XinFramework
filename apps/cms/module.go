package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/module/cms/internal/handler"
	"gx1727.com/xin/module/cms/internal/service"
)

type module struct {
	name string
}

func (m *module) Name() string { return m.name }

func (m *module) Init() error { return nil }

func (m *module) Shutdown() error { return nil }

func (m *module) Register(public, protected *gin.RouterGroup) {
	svc := service.NewService()
	h := handler.NewHandler(svc)
	Register(h, public, protected)
}

func Module() plugin.Module {
	return &module{name: "cms"}
}

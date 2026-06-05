package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

type module struct {
	name string
}

func (m *module) Name() string    { return m.name }
func (m *module) Init() error     { return nil }
func (m *module) Shutdown() error { return nil }

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	// 创建 Handler（直接调用 pkg/repository 提供的核心接口）
	h := NewHandler()
	Register(h, public, protected)
}

func Module() plugin.Module {
	return &module{name: "cms"}
}

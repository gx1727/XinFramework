package flag

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

type module struct {
	name string
}

func (m *module) Name() string { return m.name }
func (m *module) Init() error {
	return nil
}
func (m *module) Shutdown() error { return nil }

func (m *module) Register(public, protected *gin.RouterGroup) {
	// 初始化 Repository
	InitRepositories(db.Get())

	// 创建 Handler（直接调用 Repository，无 Service 层）
	h := NewHandler()
	Register(public, protected, h)
}

func Module() plugin.Module {
	return &module{name: "flag"}
}

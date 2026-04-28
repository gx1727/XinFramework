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
	frameRepo := NewFrameRepository(db.Get())
	avatarRepo := NewAvatarRepository(db.Get())
	frameCatRepo := NewFrameCategoryRepository(db.Get())
	avatarCatRepo := NewAvatarCategoryRepository(db.Get())
	svc := NewService(nil, frameRepo, avatarRepo, frameCatRepo, avatarCatRepo)
	h := NewHandler(svc)
	Register(public, protected, h)
}

func Module() plugin.Module {
	return &module{name: "flag"}
}

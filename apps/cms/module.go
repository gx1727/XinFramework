package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/module/cms/internal/handler"
	"gx1727.com/xin/module/cms/internal/service"
)

var (
	cmsService *service.Service
	cmsHandler *handler.Handler
)

type module struct {
	name string
}

func (m *module) Name() string { return m.name }

func (m *module) Init() error { return nil }

func (m *module) Shutdown() error { return nil }

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	if cmsHandler != nil {
		Register(cmsHandler, public, protected)
	}
}

// InitService 由 framework 调用，注入 Framework 的 Repository 依赖
// CMS 自己的表数据（如 CmsPost）直接使用 db.Get() 访问
func InitService(
	userRepo model.UserRepository,
	tenantRepo model.TenantRepository,
) {
	cmsService = service.NewService(userRepo, tenantRepo)
	cmsHandler = handler.NewHandler(cmsService)
}

func Module() plugin.Module {
	return &module{name: "cms"}
}

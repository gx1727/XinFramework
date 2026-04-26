package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/repository"
	"gx1727.com/xin/module/cms/internal/handler"
	"gx1727.com/xin/module/cms/internal/service"
)

type Handler = handler.Handler

func Register(h *Handler, public, protected *gin.RouterGroup) {
	public.GET("/cms/ping", h.Ping)
	protected.GET("/cms/me", h.GetCurrentUser)
	protected.GET("/cms/users", h.ListUsers)
	protected.GET("/cms/tenant", h.GetTenant)

	// CMS Posts CRUD
	protected.GET("/cms/posts", h.ListPosts)
	protected.GET("/cms/posts/:id", h.GetPost)
	protected.POST("/cms/posts", h.CreatePost)
	protected.PUT("/cms/posts/:id", h.UpdatePost)
	protected.DELETE("/cms/posts/:id", h.DeletePost)
}

type module struct {
	name string
}

func (m *module) Name() string { return m.name }

func (m *module) Init() error { return nil }

func (m *module) Shutdown() error { return nil }

func (m *module) Register(public, protected *gin.RouterGroup) {
	svc := service.NewService(
		repository.User(),
		repository.Tenant(),
		repository.CmsPost(),
	)
	h := handler.NewHandler(svc)
	Register(h, public, protected)
}

// Module creates the CMS plugin module
func Module() plugin.Module {
	return &module{name: "cms"}
}

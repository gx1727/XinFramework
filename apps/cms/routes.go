package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/module/cms/internal/handler"
)

type Handler = handler.Handler

func Register(h *Handler, public, protected *gin.RouterGroup) {
	public.GET("/cms/ping", h.Ping)
	protected.GET("/cms/me", h.GetCurrentUser)
	protected.GET("/cms/users", h.ListUsers)
	protected.GET("/cms/tenant", h.GetTenant)

	protected.GET("/cms/posts", h.ListPosts)
	protected.GET("/cms/posts/:id", h.GetPost)
	protected.POST("/cms/posts", h.CreatePost)
	protected.PUT("/cms/posts/:id", h.UpdatePost)
	protected.DELETE("/cms/posts/:id", h.DeletePost)
}

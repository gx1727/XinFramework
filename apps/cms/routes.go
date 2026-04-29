package cms

import (
	"github.com/gin-gonic/gin"
)

func Register(h *Handler, public *gin.RouterGroup, protected *gin.RouterGroup) {
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

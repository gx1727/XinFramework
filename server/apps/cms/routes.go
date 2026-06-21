package cms

import (
	"github.com/gin-gonic/gin"
)

func Register(h *Handler, public *gin.RouterGroup, tenant *gin.RouterGroup) {
	public.GET("/cms/ping", h.Ping)
	tenant.GET("/cms/me", h.GetCurrentUser)
	tenant.GET("/cms/users", h.ListUsers)
	tenant.GET("/cms/tenant", h.GetTenant)

	tenant.GET("/cms/posts", h.ListPosts)
	tenant.GET("/cms/posts/:id", h.GetPost)
	tenant.POST("/cms/posts", h.CreatePost)
	tenant.PUT("/cms/posts/:id", h.UpdatePost)
	tenant.DELETE("/cms/posts/:id", h.DeletePost)
}

package auth

import (
	"github.com/gin-gonic/gin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.POST("/auth/login", h.Login)
	public.POST("/auth/register", h.Register)
	public.POST("/auth/refresh", h.Refresh)

	protected.POST("/auth/logout", h.Logout)
}

package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterV1(public *gin.RouterGroup, protected *gin.RouterGroup) {
	h := NewHandler()
	public.POST("/login", h.Login)
	public.POST("/register", h.Register)
	protected.POST("/logout", h.Logout)
}

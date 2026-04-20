package auth

import "github.com/gin-gonic/gin"

// RegisterV1 registers auth routes for v1.
func RegisterV1(r *gin.RouterGroup) {
	h := NewHandler()
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
}

package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/internal/module/auth"
	"gx1727.com/xin/pkg/resp"
)

func RegisterRoutes(r *gin.Engine) {
	authHandler := auth.NewHandler()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			resp.Success(c, gin.H{"status": "ok"})
		})
		v1.POST("/login", authHandler.Login)
		v1.POST("/logout", authHandler.Logout)
	}
}

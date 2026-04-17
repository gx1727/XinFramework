package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xin-framework/xin/pkg/resp"
)

func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			resp.Success(c, gin.H{"status": "ok"})
		})
	}
}

package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin-framework/pkg/resp"
)

func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			resp.Success(c, gin.H{"status": "ok"})
		})
	}
}

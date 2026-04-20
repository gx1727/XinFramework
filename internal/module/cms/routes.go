package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/pkg/resp"
)

// RegisterV1 registers CMS routes for v1.
func RegisterV1(r *gin.RouterGroup) {
	r.GET("/cms/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{"domain": "cms", "status": "enabled"})
	})
}

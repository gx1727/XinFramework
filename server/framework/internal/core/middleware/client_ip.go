package middleware

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/audit"
)

// ClientIP 把请求客户端 IP 注入 ctx，供 audit.Log 写 db_logs.ip。
// 放在 RequestID 之后、CORS 之后；位置不敏感，只要在 OptionalAuth/Auth 之前即可。
func ClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := audit.WithIP(c.Request.Context(), c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

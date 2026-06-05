package middleware

import (
	"github.com/gin-gonic/gin"
)

// Recovery 异常恢复中间件 - 捕获 panic 并恢复
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// RateLimit 限流中间件 - 目前为占位实现
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gx1727.com/xin/framework/pkg/logger"
)

// RequestID 请求ID中间件 - 生成或传递 X-Request-ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger 请求日志中间件 - 记录请求信息和响应状态
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		requestID, _ := c.Get("request_id")
		reqID, _ := requestID.(string)
		if reqID == "" {
			reqID = "-"
		}

		if raw != "" {
			path = path + "?" + raw
		}

		switch {
		case statusCode >= 500:
			logger.Errorf("[%s] %s %s | %d | %v | %s", reqID, method, path, statusCode, latency, clientIP)
		case statusCode >= 400:
			logger.Warnf("[%s] %s %s | %d | %v | %s", reqID, method, path, statusCode, latency, clientIP)
		default:
			logger.Infof("[%s] %s %s | %d | %v | %s", reqID, method, path, statusCode, latency, clientIP)
		}
	}
}

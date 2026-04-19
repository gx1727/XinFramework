package middleware

import (
	"gx1727.com/xin/internal/infra/logger"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/internal/core/context"
	"gx1727.com/xin/internal/infra/db"
	"gx1727.com/xin/pkg/config"
	"gx1727.com/xin/pkg/resp"
)

func Auth(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			resp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			resp.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if userID, ok := claims["user_id"].(float64); ok {
			ctx := context.New(c)
			ctx.SetUserID(uint(userID))
		}

		c.Next()
	}
}

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

func Tenant(mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mode == "" {
			c.Next()
			return
		}

		ctx := context.New(c)
		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			if tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64); err == nil {
				ctx.SetTenantID(uint(tenantID))
				db.SetTenantID(uint(tenantID))
			}
		}

		defer db.ClearTenantID()
		c.Next()
	}
}

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

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

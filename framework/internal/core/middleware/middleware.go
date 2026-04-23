package middleware

import (
	"strconv"
	"strings"
	"time"

	"gx1727.com/xin/framework/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gx1727.com/xin/framework/internal/core/context"
	"gx1727.com/xin/framework/pkg/config"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/framework/pkg/session"
)

func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	if cfg == nil || !cfg.Enabled || len(cfg.AllowOrigins) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowOrigin := ""
		for _, o := range cfg.AllowOrigins {
			o = strings.TrimSpace(o)
			if o == "*" {
				allowOrigin = "*"
				break
			}
			if strings.EqualFold(o, origin) {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin == "" {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)
		c.Header("Access-Control-Allow-Headers", cfg.AllowHeaders)
		c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func Auth(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			resp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &jwtpkg.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			resp.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		ok, err := session.Validate(claims.SessionID)
		if err != nil || !ok {
			resp.Unauthorized(c, "session expired or revoked")
			c.Abort()
			return
		}

		ctx := context.New(c)
		ctx.SetUserID(claims.UserID)
		ctx.SetTenantID(claims.TenantID)
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("session_id", claims.SessionID)
		c.Set("role", claims.Role)

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
				tid := uint(tenantID)
				ctx.SetTenantID(tid)
				c.Set("tenant_id", tid)
				c.Request = c.Request.WithContext(context.WithTenantID(c.Request.Context(), tid))
			}
		}

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

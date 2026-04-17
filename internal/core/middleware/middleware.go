package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xin-framework/xin/configs"
	"github.com/xin-framework/xin/internal/core/context"
)

func Auth(cfg *configs.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid token"})
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

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		ctx := context.New(c)
		if tenantID != "" {
			var tid uint
			if _, err := gin.DefaultBinding.BindTenantParams(c, &tid); err == nil {
				ctx.SetTenantID(tid)
			}
		}
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return gin.Logger()
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

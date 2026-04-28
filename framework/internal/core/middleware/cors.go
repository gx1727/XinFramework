package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/config"
)

// CORS 跨域资源共享中间件
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

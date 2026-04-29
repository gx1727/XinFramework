package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	xinContext "gx1727.com/xin/framework/pkg/context"
)

// Tenant 租户隔离中间件 - 从请求头获取租户ID并设置到上下文
func Tenant(mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mode == "" {
			c.Next()
			return
		}

		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			if tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64); err == nil {
				tid := uint(tenantID)

				var xc *xinContext.XinContext
				if existingXc, ok := xinContext.XinContextFrom(c.Request.Context()); ok {
					xc = existingXc.Clone()
					xc.TenantID = tid
				} else {
					xc = &xinContext.XinContext{TenantID: tid}
				}

				c.Request = c.Request.WithContext(xinContext.WithTenantID(xinContext.WithXinContext(c.Request.Context(), xc), tid))
				c.Set("tenant_id", tid)
			}
		}

		c.Next()
	}
}

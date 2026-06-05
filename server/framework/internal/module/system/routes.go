package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})

	system := protected.Group("/system")
	{
		system.GET("/server-info", middleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.ServerInfo)
		system.POST("/clear-cache", middleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.ClearCache)

		// Redis Cache Controls
		system.GET("/cache/info", middleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.CacheInfo)
		system.GET("/cache/keys", middleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheKeys)
		system.GET("/cache/value/*key", middleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheValue)
		system.DELETE("/cache/keys/*key", middleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.DeleteCacheKey)
	}
}

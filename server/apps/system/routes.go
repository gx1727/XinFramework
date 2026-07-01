package system

import (
	"github.com/gin-gonic/gin"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
)

// Register 注册 system 路由。
//   - public:    /health                          （公开）
//   - tenant:    /system/server-info              （业务域运维，只读，租户可见）
//   - protected: /sys/system/cache/*              （sys 域 Redis cache 运维，需 super_admin）
//     /sys/system/clear-cache
func Register(public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})

	system := tenant.Group("/system")
	{
		system.GET("/server-info", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.ServerInfo)
	}

	// Redis Cache Controls：仅 super_admin（sys 域）
	sysSystem := protected.Group("/system")
	{
		sysSystem.POST("/clear-cache", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.ClearCache)

		sysSystem.GET("/cache/info", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.CacheInfo)
		sysSystem.GET("/cache/keys", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheKeys)
		sysSystem.GET("/cache/value/*key", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheValue)
		sysSystem.DELETE("/cache/keys/*key", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.DeleteCacheKey)
	}
}

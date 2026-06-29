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
//   - protected: /platform/system/cache/*         （平台域 Redis cache 运维，需 super_admin）
//                /platform/system/clear-cache
func Register(public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})

	system := tenant.Group("/system")
	{
		system.GET("/server-info", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.ServerInfo)
	}

	// Redis Cache Controls：仅 super_admin（platform 域）
	platformSystem := protected.Group("/system")
	{
		platformSystem.POST("/clear-cache", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.ClearCache)

		platformSystem.GET("/cache/info", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.CacheInfo)
		platformSystem.GET("/cache/keys", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheKeys)
		platformSystem.GET("/cache/value/*key", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheValue)
		platformSystem.DELETE("/cache/keys/*key", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.DeleteCacheKey)
	}
}

package system

import (
	"github.com/gin-gonic/gin"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
)

// Register 注册 system 路由。
//   - public:    /health              （公开）
//   - tenant:    /system/*            （业务域运维，需 tenant）
//   - protected: /platform/system/*   （平台域运维，需 super_admin，暂留空）
func Register(public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})

	system := tenant.Group("/system")
	{
		system.GET("/server-info", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.ServerInfo)
		system.POST("/clear-cache", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.ClearCache)

		// Redis Cache Controls
		system.GET("/cache/info", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.CacheInfo)
		system.GET("/cache/keys", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheKeys)
		system.GET("/cache/value/*key", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActList)), h.GetCacheValue)
		system.DELETE("/cache/keys/*key", pkgmiddleware.Require(permission.P(permission.ResSystem, permission.ActUpdate)), h.DeleteCacheKey)
	}
}

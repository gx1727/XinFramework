package system

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/resp"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	public.GET("/health", func(c *gin.Context) {
		resp.Success(c, gin.H{"status": "ok"})
	})

	system := protected.Group("/system")
	{
		system.GET("/server-info", middleware.RequirePermission("system", "list"), h.ServerInfo)
		system.POST("/clear-cache", middleware.RequirePermission("system", "update"), h.ClearCache)
	}
}

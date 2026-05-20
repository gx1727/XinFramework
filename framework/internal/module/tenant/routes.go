package tenant

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	tenants := protected.Group("/tenants")
	{
		tenants.POST("", middleware.RequirePermission(permission.ResTenant, permission.ActCreate), h.Create)
		tenants.PUT("/:id", middleware.RequirePermission(permission.ResTenant, permission.ActUpdate), h.Update)
		tenants.DELETE("/:id", middleware.RequirePermission(permission.ResTenant, permission.ActDelete), h.Delete)
		tenants.GET("/:id", middleware.RequirePermission(permission.ResTenant, permission.ActList), h.Get)
		tenants.GET("", middleware.RequirePermission(permission.ResTenant, permission.ActList), h.List)
	}
}

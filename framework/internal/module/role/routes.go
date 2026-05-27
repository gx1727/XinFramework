package role

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	roles := protected.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission(permission.ResRole, permission.ActList), h.List)
		roles.GET("/:id", middleware.RequirePermission(permission.ResRole, permission.ActList), h.Get)
		roles.POST("", middleware.RequirePermission(permission.ResRole, permission.ActCreate), h.Create)
		roles.PUT("/:id", middleware.RequirePermission(permission.ResRole, permission.ActUpdate), h.Update)
		roles.DELETE("/:id", middleware.RequirePermission(permission.ResRole, permission.ActDelete), h.Delete)
		roles.GET("/:id/data-scopes", middleware.RequirePermission(permission.ResRole, permission.ActList), h.GetDataScopes)
		roles.PUT("/:id/data-scopes", middleware.RequirePermission(permission.ResRole, permission.ActUpdate), h.UpdateDataScopes)
		// 角色菜单权限
		roles.GET("/:id/menus", middleware.RequirePermission(permission.ResRole, permission.ActList), h.GetMenus)
		roles.PUT("/:id/menus", middleware.RequirePermission(permission.ResRole, permission.ActUpdate), h.AssignMenus)
	}
}

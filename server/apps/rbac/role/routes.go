package role

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(tenant *gin.RouterGroup, h *Handler) {
	roles := tenant.Group("/roles")
	{
		roles.GET("", middleware.Require(permission.P(permission.ResRole, permission.ActList)), h.List)
		roles.GET("/:id", middleware.Require(permission.P(permission.ResRole, permission.ActList)), h.Get)
		roles.POST("", middleware.Require(permission.P(permission.ResRole, permission.ActCreate)), h.Create)
		roles.PUT("/:id", middleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.Update)
		roles.PATCH("/:id", middleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.Patch)
		roles.DELETE("/:id", middleware.Require(permission.P(permission.ResRole, permission.ActDelete)), h.Delete)
		roles.GET("/:id/data-scopes", middleware.Require(permission.P(permission.ResRole, permission.ActList)), h.GetDataScopes)
		roles.PUT("/:id/data-scopes", middleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.UpdateDataScopes)
		// 角色菜单权限
		roles.GET("/:id/menus", middleware.Require(permission.P(permission.ResRole, permission.ActList)), h.GetMenus)
		roles.PUT("/:id/menus", middleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.AssignMenus)
	}
}

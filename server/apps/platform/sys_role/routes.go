package sysrole

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// 0024+：删除 RequirePlatformRole(super_admin) 硬编码白名单。
// 任何 platform 角色都可以调到这里；具体能力由 ResRole:* 资源权限码决定。
// super_admin 靠 init_seed.sql 11.3c 绑定的 `*:*` 通配自动拥有。
func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/sys-roles",
		pkgmiddleware.RequireAnyPlatformRole(),
	)
	{
		g.GET("", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActList)), h.List)
		g.POST("", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActCreate)), h.Create)
		g.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActGet)), h.Get)
		g.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.Update)
		g.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActDelete)), h.Delete)
		g.PUT("/:id/menus", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.AssignMenus)
		g.PUT("/:id/permissions", pkgmiddleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.AssignPermissions)
	}
}

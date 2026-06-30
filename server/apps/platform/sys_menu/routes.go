package sysmenu

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// 0024+：删除 RequirePlatformRole(super_admin) 硬编码白名单。
// 任何 platform 角色都可以调到 CRUD；具体能力由 ResMenu:* 资源权限码决定。
// 运行时 Tree 仍对任何 platform 角色开放（每个角色按 sys_role_menus 收敛）。
// super_admin 靠 init_seed.sql 11.3c 绑定的 `*:*` 通配自动拥有 CRUD 权限。
func Register(protected *gin.RouterGroup, h *Handler) {
	// 平台菜单 CRUD
	adminG := protected.Group("/menus",
		pkgmiddleware.RequireAnyPlatformRole(),
	)
	{
		adminG.GET("", pkgmiddleware.Require(permission.P(permission.ResMenu, permission.ActList)), h.List)
		adminG.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResMenu, permission.ActGet)), h.Get)
		adminG.POST("", pkgmiddleware.Require(permission.P(permission.ResMenu, permission.ActCreate)), h.Create)
		adminG.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResMenu, permission.ActUpdate)), h.Update)
		adminG.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResMenu, permission.ActDelete)), h.Delete)
	}

	// 平台菜单树（运行时）：任何 platform 角色均可访问。
	// service 层按调用者被分配的 sys_role_menus 收敛（替换旧 isSuperAdmin 分支）。
	runtimeG := protected.Group("/menus",
		pkgmiddleware.RequireAnyPlatformRole(),
	)
	{
		runtimeG.GET("/tree", h.Tree)
	}
}

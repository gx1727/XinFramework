package sysmenu

import (
	"github.com/gin-gonic/gin"

	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	// 平台菜单 CRUD 仅 super_admin。
	adminG := protected.Group("/menus",
		pkgmiddleware.RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin),
	)
	{
		adminG.GET("", h.List)
		adminG.GET("/:id", h.Get)
		adminG.POST("", h.Create)
		adminG.PUT("/:id", h.Update)
		adminG.DELETE("/:id", h.Delete)
	}

	// 平台菜单树（运行时）：任何 platform 角色均可访问，
	// handler 按调用者的 PlatformRoles 在 SQL 层过滤，
	// super_admin 仍然全量。
	//
	// 为什么 Tree 与 CRUD 分开：同一份菜单数据在两种场景下的“可见性语义”不同。
	//  1. CRUD：所有平台菜单的“管理面”列表，super_admin 需要全量才能分配授权。
	//  2. Runtime：当前登录平台用户能看到的“运营面”菜单，应按其角色收敛。
	runtimeG := protected.Group("/menus",
		pkgmiddleware.RequireAnyPlatformRole(),
	)
	{
		runtimeG.GET("/tree", h.Tree)
	}
}

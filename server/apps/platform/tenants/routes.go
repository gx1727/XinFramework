package tenants

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// PlatformRoleSuperAdmin 平台级超级管理员角色名。租户管理属于跨租户特权：
// 仅允许持有该平台角色的账号访问　
const PlatformRoleSuperAdmin = "super_admin"

// Register 把平台租户管理路由挂刀/api/v1/platform/tenants　
//
// 路径约定（与 sys_menu 等 platform 域模块一致）：
//   - /platform 子空间表礀平台管理埀
//   - /tenants 直接挂资源（旀/platform-tenants 这层嵌套：
//
// 中间件顺序：
//  1. protected.Use(middleware.Auth(...))            // 来自 framework.go：注兀XinContext
//  2. g := protected.Group("/tenants",
//                       pkgmiddleware.RequirePlatformRole("super_admin"))
//  3. 各路由上叠加 pkgmiddleware.Require(permission.P(...)) 做资源级权限细分
//
// 即使持有 super_admin，仍需满足资源权限码（tenant:create / update / delete / list）　
// 两个守卫都过才算合法——避免任佀tenant admin 仅凭资源权限码越权　
func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/tenants",
		pkgmiddleware.RequirePlatformRole(PlatformRoleSuperAdmin),
	)
	{
		g.POST("", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActCreate)), h.Create)
		g.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActUpdate)), h.Update)
		g.PUT("/:id/status", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActUpdate)), h.UpdateStatus)
		g.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActDelete)), h.Delete)
		// Purge 单独 endpoint：硬删是不可逆操作，URL 用动词区分，避免误用 DELETE 触发　
		g.POST("/:id/purge", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActDelete)), h.Purge)
		g.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.Get)
		g.GET("", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.List)
		// Impersonate 模拟登录：高敏操作，复用 ResTenant:list 资源权限
		// （super_admin 平台角色守卫已在 group 级；资源权限保持最宽松）
		g.POST("/:id/impersonate", pkgmiddleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.Impersonate)
	}
}
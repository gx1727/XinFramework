package sysuser

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// Register 把 sys_user 路由挂到 /api/v1/platform/sys-users。
//
// 中间件链：
//   - protected.Use(middleware.Auth(...))      来自 framework.go（已挂）
//   - group 级 RequireAnyPlatformRole：挡住非平台用户
//   - 各 route 叠加 Require(P(ResUser, ...))：按 RBAC 决定具体能力
//
// 0024+：删除 RequirePlatformRole(super_admin) 硬编码白名单。
// 任何 platform 角色都可以调到这里；能不能创建/修改/删除，取决于该角色是否
// 拥有对应的 ResUser:* 资源权限码。super_admin 靠 init_seed.sql 11.3c 绑定的
// `*:*` 通配自动拥有所有资源权限。
func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/sys-users",
		pkgmiddleware.RequireAnyPlatformRole(),
	)
	{
		g.GET("", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActList)), h.List)
		g.POST("", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActCreate)), h.Create)
		g.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActGet)), h.Get)
		g.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.Update)
		g.PUT("/:id/status", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.UpdateStatus)
		g.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActDelete)), h.Delete)
		g.PUT("/:id/roles", pkgmiddleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.AssignRoles)
	}
}

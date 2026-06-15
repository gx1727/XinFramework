package tenant

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// PlatformRoleSuperAdmin 平台级超级管理员角色名。租户管理属于跨租户特权，
// 仅允许持有该平台角色的账号访问。
const PlatformRoleSuperAdmin = "super_admin"

func Register(protected *gin.RouterGroup, h *Handler) {
	// 平台守卫：先校验 super_admin 平台角色，再叠加资源级 RBAC。
	// 两个守卫都过才算合法——避免任何租户内 admin 仅凭资源权限码越权。
	tenants := protected.Group("/tenants")
	tenants.Use(middleware.RequirePlatformRole(PlatformRoleSuperAdmin))
	{
		tenants.POST("", middleware.Require(permission.P(permission.ResTenant, permission.ActCreate)), h.Create)
		tenants.PUT("/:id", middleware.Require(permission.P(permission.ResTenant, permission.ActUpdate)), h.Update)
		tenants.PUT("/:id/status", middleware.Require(permission.P(permission.ResTenant, permission.ActUpdate)), h.UpdateStatus)
		tenants.DELETE("/:id", middleware.Require(permission.P(permission.ResTenant, permission.ActDelete)), h.Delete)
		// Purge 单独 endpoint：硬删是不可逆操作，URL 用动词区分，避免误用 DELETE 触发。
		tenants.POST("/:id/purge", middleware.Require(permission.P(permission.ResTenant, permission.ActDelete)), h.Purge)
		tenants.GET("/:id", middleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.Get)
		tenants.GET("", middleware.Require(permission.P(permission.ResTenant, permission.ActList)), h.List)
	}
}

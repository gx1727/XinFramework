package permission

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// Register 注册权限管理路由
// 资源权限（按钮/API）通过 role_resources 表管理
// 菜单权限已迁移到 role 模块（/roles/:id/menus）
func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/roles/:id/permissions", middleware.Require(permission.P(permission.ResRole, permission.ActList)), h.GetPermissions)
	protected.POST("/roles/:id/permissions", middleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.AssignPermissions)
	protected.PUT("/roles/:id/permissions", middleware.Require(permission.P(permission.ResRole, permission.ActUpdate)), h.AssignPermissions)
	protected.GET("/roles/:id/resources", middleware.Require(permission.P(permission.ResRole, permission.ActList)), h.GetResources)
}

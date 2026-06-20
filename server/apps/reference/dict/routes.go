// Package dict 字典路由注册
package dict

import (
	"github.com/gin-gonic/gin"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// PlatformRoleSuperAdmin 平台级 super_admin 角色名（与 tenant 模块保持一致）
const PlatformRoleSuperAdmin = "super_admin"

func Register(protected *gin.RouterGroup, h *Handler) {
	// ============ 业务消费入口（所有登录用户可访问） ============
	d := protected.Group("/dicts")
	{
		// 合并字典：业务最终消费（租户视角）
		d.GET("/resolve", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.Resolve)
		d.POST("/resolve/batch", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.ResolveBatch)

		// 租户自建字典（原有 API）
		d.GET("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.List)
		d.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.Get)
		d.POST("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.Create)
		d.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.Update)
		d.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.Delete)

		// 字典项：写在同一资源下
		d.GET("/:id/items", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListItems)
		d.POST("/:id/items", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.CreateItem)
		d.PUT("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpdateItem)
		d.DELETE("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteItem)

		// 租户覆盖：POST/DELETE 单条覆盖
		d.PUT("/:id/items/:item_id/override", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpsertOverride)
		d.DELETE("/:id/items/:item_id/override", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteOverride)
	}

	// ============ super_admin：平台字典 CRUD ============
	//     先用 RequirePlatformRole 网关挡住非 super_admin，再叠加资源 RBAC
	pd := protected.Group("/dicts/platform")
	pd.Use(pkgmiddleware.RequirePlatformRole(PlatformRoleSuperAdmin))
	{
		pd.GET("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListPlatformDicts)
		pd.POST("", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.CreatePlatformDict)
		pd.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActGet)), h.GetPlatformDict)
		pd.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpdatePlatformDict)
		pd.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.DeletePlatformDict)

		// 平台字典项
		pd.GET("/:id/items", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListPlatformItems)
		pd.POST("/:id/items", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActCreate)), h.CreatePlatformItem)
		pd.PUT("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpdatePlatformItem)
		pd.DELETE("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActDelete)), h.DeletePlatformItem)

		// 可见性配置
		pd.GET("/:id/visibility", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActList)), h.ListVisibility)
		pd.POST("/:id/visibility", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.UpsertVisibility)
		pd.DELETE("/:id/visibility/:tenant_id", pkgmiddleware.Require(permission.P(permission.ResDict, permission.ActUpdate)), h.DeleteVisibility)
	}
}
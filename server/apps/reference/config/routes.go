// Package config 通用配置 - 路由注册
//
// 与 apps/reference/dict 的路由结构对齐：
//
//   /api/v1/configs                       业务消费 + 租户自建（Require ResConfig）
//     GET    /                            ListGroups
//     GET    /:id                         GetGroup（按 id 查并 resolve）
//     GET    /:id/items                   ListItemsByGroup
//     POST   /:id/items/:item_id/override UpsertOverride
//     DELETE /:id/items/:item_id/override DeleteOverride
//     GET    /resolve                     Resolve（?code=xxx 合并消费）
//     POST   /resolve/batch               ResolveBatch
//
//   /api/v1/configs/platform              super_admin 平台 CRUD
//     RequirePlatformRole("super_admin") + Require ResConfig
//     GET    /                            ListPlatformGroups
//     GET    /:id                         GetPlatformGroup
//     POST   /                            CreatePlatformGroup
//     PUT    /:id                         UpdatePlatformGroup
//     DELETE /:id                         DeletePlatformGroup
//     GET    /:id/items                   ListPlatformItems
//     POST   /:id/items                   CreatePlatformItem
//     PUT    /:id/items/:item_id          UpdatePlatformItem
//     DELETE /:id/items/:item_id          DeletePlatformItem
//     GET    /:id/visibility              ListVisibility
//     POST   /:id/visibility              UpsertVisibility
//     DELETE /:id/visibility/:tenant_id   DeleteVisibility
//
//   /api/v1/public/configs                公开读（无需 auth，仅 X-Tenant-ID header）
//     GET    /                            GetPublic
package config

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

const PlatformRoleSuperAdmin = "super_admin"

// Register 注册三组路由
func Register(public *gin.RouterGroup, protected *gin.RouterGroup, bh *BusinessHandler, ph *PlatformHandler, pubh *PublicHandler) {
	// ============ Public（无需 auth，独立前缀避免与 /configs 冲突） ============
	pub := public.Group("/public/configs")
	{
		pub.GET("", pubh.GetPublic)
	}

	// ============ Business（业务消费 + 租户自建） ============
	biz := protected.Group("/configs")
	biz.Use(pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList))) // 默认要求 list 权限
	{
		// 合并消费端点（最高频）
		biz.GET("/resolve", bh.Resolve)
		biz.POST("/resolve/batch", bh.ResolveBatch)

		// 租户视角的 group / item 查询
		biz.GET("", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), bh.ListGroups)
		biz.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActGet)), bh.GetGroup)
		biz.GET("/:id/items", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), bh.ListItemsByGroup)

		// Override（租户覆盖 platform item）
		biz.POST("/:id/items/:item_id/override", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), bh.UpsertOverride)
		biz.DELETE("/:id/items/:item_id/override", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), bh.DeleteOverride)
	}

	// ============ Platform（super_admin 平台 CRUD） ============
	plat := protected.Group("/configs/platform")
	plat.Use(pkgmiddleware.RequirePlatformRole(PlatformRoleSuperAdmin))
	{
		// Group
		plat.GET("", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), ph.ListGroups)
		plat.GET("/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActGet)), ph.GetGroup)
		plat.POST("", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActCreate)), ph.CreateGroup)
		plat.PUT("/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), ph.UpdateGroup)
		plat.DELETE("/:id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActDelete)), ph.DeleteGroup)

		// Item
		plat.GET("/:id/items", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), ph.ListItems)
		plat.POST("/:id/items", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActCreate)), ph.CreateItem)
		plat.PUT("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), ph.UpdateItem)
		plat.DELETE("/:id/items/:item_id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActDelete)), ph.DeleteItem)

		// Visibility
		plat.GET("/:id/visibility", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActList)), ph.ListVisibility)
		plat.POST("/:id/visibility", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), ph.UpsertVisibility)
		plat.DELETE("/:id/visibility/:tenant_id", pkgmiddleware.Require(permission.P(permission.ResConfig, permission.ActUpdate)), ph.DeleteVisibility)
	}
}

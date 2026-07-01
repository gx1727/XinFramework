// Package config 通用配置 - 路由注册
//
// 路由空间（重构后）：
//
//	/api/v1/configs                       业务消费 + 租户自建（Auth + RequireTenantContext + Require ResConfig）
//	  GET    /                            ListGroups
//	  GET    /:id                         GetGroup
//	  GET    /:id/items                   ListItemsByGroup
//	  POST   /:id/items/:item_id/override UpsertOverride
//	  DELETE /:id/items/:item_id/override DeleteOverride
//	  GET    /resolve                     Resolve（?code=xxx 合并消费）
//	  POST   /resolve/batch               ResolveBatch
//
//	/api/v1/sys/configs                   sys 域 CRUD
//	  RequireAnySysRole() + Require ResConfig（0024+）
//	  GET    /                            ListPlatformGroups
//	  GET    /:id                         GetPlatformGroup
//	  POST   /                            CreatePlatformGroup
//	  PUT    /:id                         UpdatePlatformGroup
//	  DELETE /:id                         DeletePlatformGroup
//	  GET    /:id/items                   ListPlatformItems
//	  POST   /:id/items                   CreatePlatformItem
//	  PUT    /:id/items/:item_id          UpdatePlatformItem
//	  DELETE /:id/items/:item_id          DeletePlatformItem
//	  GET    /:id/visibility              ListVisibility
//	  POST   /:id/visibility              UpsertVisibility
//	  DELETE /:id/visibility/:tenant_id   DeleteVisibility
//
//	/api/v1/public/configs                公开读（无需 auth，仅 X-Tenant-ID header）
//	  GET    /                            GetPublic
package config

import (
	"github.com/gin-gonic/gin"

	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

// Register 注册三组路由
//
// 三组 RouterGroup 语义：
//   - public:     /api/v1/public/configs          （公开读）
//   - tenant:     /api/v1/configs                 （业务域，Auth + RequireTenantContext）
//   - protected:  /api/v1/sys/configs             （sys 域，Auth + RequireSysRole）
func Register(public *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup, bh *BusinessHandler, ph *PlatformHandler, pubh *PublicHandler) {
	// ============ Public（无需 auth，独立前缀避免与 /configs 冲突） ============
	pub := public.Group("/public/configs")
	{
		pub.GET("", pubh.GetPublic)
	}

	// ============ Business（业务消费 + 租户自建） ============
	biz := tenant.Group("/configs")
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

	// ============ sys（sys 域 CRUD，挂在 /sys/* 域） ============
	// 0024+：删除 RequireSysRole(super_admin) 硬编码白名单。
	// 任何 sys 角色都可以调到这里；具体能力由 ResConfig:* 资源权限码决定。
	// super_admin 靠 init_seed.sql 11.3c 绑定的 `*:*` 通配自动拥有。
	plat := protected.Group("/configs")
	plat.Use(pkgmiddleware.RequireAnySysRole())
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

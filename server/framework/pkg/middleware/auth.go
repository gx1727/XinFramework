package middleware

import (
	"github.com/gin-gonic/gin"
	xinContext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
)

// Require 创建权限检查中间件 - 用户必须拥有指定权限才能访问
// 用法: protected.POST("/flag/frames", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateFrame)
func Require(spec permission.Spec) gin.HandlerFunc {
	return requireWithSpecs("one", spec)
}

// RequireAuthenticated 创建登录检查中间件 - 只需登录即可访问，不检查具体权限
// 用法: protected.GET("/profile", middleware.RequireAuthenticated(), h.GetProfile)
func RequireAuthenticated() gin.HandlerFunc {
	return requireWithSpecs("one", permission.AuthOnly())
}

// RequireAny 创建任意权限检查中间件 - 用户拥有任意一个指定权限即可访问
// 用法: protected.DELETE("/admin", middleware.RequireAny(permission.P(permission.ResUser, permission.ActDelete), permission.P(permission.ResAdmin, permission.ActDelete)), h.Delete)
func RequireAny(specs ...permission.Spec) gin.HandlerFunc {
	return requireWithSpecs("any", specs...)
}

// RequireAll 创建全部权限检查中间件 - 用户必须拥有所有指定权限才能访问
// 用法: protected.GET("/admin", middleware.RequireAll(specs...), h.List)
func RequireAll(specs ...permission.Spec) gin.HandlerFunc {
	return requireWithSpecs("all", specs...)
}

// RequirePlatformRole 校验当前登录账号是否携带指定的平台级角色（如 super_admin）。
//
// 设计意图：跨租户 / 平台级操作（如租户管理、计费管理、平台字典）必须显式校验
// 平台角色，不能仅依赖资源权限码——因为资源权限是租户内的 RBAC，无法表达
// "跨越所有租户"的特权。
//
// 注意：该中间件依赖 Auth 中间件先注入 XinContext.PlatformRoles。
// 使用方式：在 protected 路由分组之后链式追加，或在单条路由上叠加。
//
// 此函数从 framework/internal/core/middleware/auth.go 提升而来，
// 因为 apps/boot/tenant 等外部业务模块需要使用它，而 internal/
// 不允许跨 module 导入。
// RequireTenantContext 校验当前 token 携带有效的 tenant_id（> 0）。
//
// 用途：挂在租户业务域路由组（如 /api/v1/t/*）上，挡住 platform 域登录的 token。
// 错误码 3003（"租户上下文缺失"）。
//
// 注意：必须挂在 Auth 中间件之后（依赖 XinContext.TenantID）。
func RequireTenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		xc := xinContext.New(c)
		if xc == nil || xc.TenantID == 0 {
			resp.Forbidden(c, "此接口要求租户上下文，请使用租户域登录")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePlatformScope 校验当前 token 是 platform 域登录（scope=platform）。
//
// 用途：挂在平台域路由组上，挡住 tenant 域登录的 token。
// 配合 RequirePlatformRole("super_admin") 使用效果最佳。
func RequirePlatformScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		xc := xinContext.New(c)
		if xc == nil || xc.TenantID != 0 {
			resp.Forbidden(c, "此接口仅限平台域登录访问")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePlatformRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(roles) == 0 {
			c.Next()
			return
		}
		xc := xinContext.New(c)
		if xc == nil || len(xc.PlatformRoles) == 0 {
			resp.Forbidden(c, "需要平台级角色")
			c.Abort()
			return
		}
		for _, need := range roles {
			for _, have := range xc.PlatformRoles {
				if have == need {
					c.Next()
					return
				}
			}
		}
		resp.Forbidden(c, "平台角色不足")
		c.Abort()
	}
}

func requireWithSpecs(mode string, specs ...permission.Spec) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc := xinContext.MustNewUserContext(c)

		// 平台超级管理员：无视所有权限规格直接放行
		if uc.IsSuperAdmin() {
			c.Next()
			return
		}

		switch mode {
		case "one":
			spec := specs[0]
			if !spec.IsValid() {
				resp.Forbidden(c, "invalid permission spec")
				c.Abort()
				return
			}
			if spec.IsAuthOnly() {
				c.Next()
				return
			}
			if !uc.HasPermission(spec.Resource, spec.Action) {
				resp.Forbidden(c, "permission denied: "+spec.String())
				c.Abort()
				return
			}
		case "any":
			for _, spec := range specs {
				if !spec.IsValid() {
					continue
				}
				if spec.IsAuthOnly() || uc.HasPermission(spec.Resource, spec.Action) {
					c.Next()
					return
				}
			}
			resp.Forbidden(c, "permission denied")
			c.Abort()
			return
		case "all":
			for _, spec := range specs {
				if !spec.IsValid() {
					resp.Forbidden(c, "invalid permission spec")
					c.Abort()
					return
				}
				if spec.IsAuthOnly() {
					continue
				}
				if !uc.HasPermission(spec.Resource, spec.Action) {
					resp.Forbidden(c, "permission denied: "+spec.String())
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

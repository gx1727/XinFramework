package middleware

import (
	"github.com/gin-gonic/gin"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/framework/pkg/xincontext"
)

// Require 创建权限检查中间件 - 用户必须拥有指定权限才能访问
// 用法: protected.POST("/flag/frames", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateFrame)
func Require(spec permission.Spec) gin.HandlerFunc {
	return requireWithSpecs(permission.MatchAll, spec)
}

// RequireAuthenticated 创建登录检查中间件 - 只需登录即可访问，不检查具体权限
// 用法: protected.GET("/profile", middleware.RequireAuthenticated(), h.GetProfile)
func RequireAuthenticated() gin.HandlerFunc {
	return requireWithSpecs(permission.MatchAll, permission.AuthOnly())
}

// RequireAny 创建任意权限检查中间件 - 用户拥有任意一个指定权限即可访问
// 用法: protected.DELETE("/admin", middleware.RequireAny(permission.P(permission.ResUser, permission.ActDelete), permission.P(permission.ResAdmin, permission.ActDelete)), h.Delete)
func RequireAny(specs ...permission.Spec) gin.HandlerFunc {
	return requireWithSpecs(permission.MatchAny, specs...)
}

// RequireAll 创建全部权限检查中间件 - 用户必须拥有所有指定权限才能访问
// 用法: protected.GET("/admin", middleware.RequireAll(specs...), h.List)
func RequireAll(specs ...permission.Spec) gin.HandlerFunc {
	return requireWithSpecs(permission.MatchAll, specs...)
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
// 用途：挂在租户业务域路由组（如 /api/v1/* 业务域）上，挡住 platform 域登录的 token。
// 错误码 3003（"租户上下文缺失"）。
//
// 注意：必须挂在 Auth 中间件之后（依赖 XinContext.TenantID）。
func RequireTenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		xc := xincontext.New(c)
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
		xc := xincontext.New(c)
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
		xc := xincontext.New(c)
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

// RequireAnyPlatformRole 校验当前 token 至少携带一个平台级角色（不限定具体角色）。
//
// 适用场景：平台域运行时接口（如 /platform/menus/tree 给任何 platform 用户拉自己的
// 可访问菜单树）。该中间件仅检查“是不是平台用户”，具体能看哪些资源由
// handler/service 层按 PlatformRoles 过滤实现。
//
// 与 RequirePlatformRole 的区别：
//   - RequirePlatformRole(super_admin) 严格白名单，仅放行指定角色
//   - RequireAnyPlatformRole()          宽泛闸口，放行任何 platform 角色
//   - RequirePlatformRole()             【注意】零参调用会被原函数视为
//     “不指定角色”直接放行，并不会检查 PlatformRoles 长度——不要用它代替本函数。
func RequireAnyPlatformRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		xc := xincontext.New(c)
		if xc == nil || len(xc.PlatformRoles) == 0 {
			resp.Forbidden(c, "需要平台级角色")
			c.Abort()
			return
		}
		c.Next()
	}
}

func requireWithSpecs(mode permission.MatchMode, specs ...permission.Spec) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc, ok := xincontext.UserContextFrom(c.Request.Context())
		if !ok || uc == nil {
			// 没有 UserContext 是路由配置错误（中间件链漏挂 Auth），
			// 用 500 + 显式 message 比 panic 友好；前端能看到，
			// SRE 能在日志里直接定位。
			resp.ServerError(c, "missing user context: did the Auth middleware run?")
			c.Abort()
			return
		}

		// 平台超级管理员：无视所有权限规格直接放行
		if uc.HasPlatformRole(jwtpkg.PlatformRoleSuperAdmin) {
			c.Next()
			return
		}

		switch mode {
		case permission.MatchAll:
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
		case permission.MatchAny:
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
		}

		c.Next()
	}
}

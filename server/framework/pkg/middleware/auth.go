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

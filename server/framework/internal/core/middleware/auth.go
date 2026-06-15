package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/framework/pkg/config"
	xinContext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/framework/pkg/session"
)

// SecurityContextLoader defines the authorization methods needed by auth middleware.
type SecurityContextLoader interface {
	LoadUserSecurityContext(ctx context.Context, userID uint) (map[string]bool, []string, *permission.DataScope, int64, error)
}

// processAuthToken extracts and validates the token, returning claims if successful
func processAuthToken(c *gin.Context, cfg *config.JWTConfig, sm session.SessionManager) (*jwtpkg.Claims, error) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return nil, jwt.ErrTokenUnverifiable
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	claims := &jwtpkg.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	ok, err := sm.Validate(claims.SessionID)
	if err != nil || !ok {
		return nil, jwt.ErrTokenExpired
	}

	return claims, nil
}

// injectAuthContext loads permissions and injects UserContext and XinContext into the request
func injectAuthContext(c *gin.Context, claims *jwtpkg.Claims, permSvc SecurityContextLoader) {
	ctx := c.Request.Context()

	// 始终优先装配轻量的 XinContext（包含了身份的基本标识）
	var xc *xinContext.XinContext
	if existingXc, ok := xinContext.XinContextFrom(ctx); ok {
		xc = existingXc.Clone()
		xc.TenantID = claims.TenantID
		xc.UserID = claims.UserID
		xc.SessionID = claims.SessionID
		xc.Role = claims.Role
		xc.PlatformRoles = claims.PlatformRoles
	} else {
		xc = &xinContext.XinContext{
			TenantID:      claims.TenantID,
			UserID:        claims.UserID,
			SessionID:     claims.SessionID,
			Role:          claims.Role,
			PlatformRoles: claims.PlatformRoles,
		}
	}
	ctx = xinContext.WithXinContext(ctx, xc)
	ctx = xinContext.WithTenantID(ctx, claims.TenantID)

	// 注册懒加载生成器到 Context 的某个钩子里，
	// 当实际业务中有人调用 MustNewUserContext 时才去查 DB 构建 UserContext
	ctx = xinContext.WithUserContextLoader(ctx, func() *xinContext.UserContext {
		var roles []string
		var perms map[string]bool
		var ds permission.DataScope
		var orgID int64

		if permSvc != nil {
			var dsPtr *permission.DataScope
			// 因为 RLS 策略强制要求 tenant_id，我们需要包裹在租户事务中
			// 只有设置了 app.tenant_id 才能查询出 users/roles/permissions
			err := db.RunInTenantTx(ctx, db.Get(), claims.TenantID, func(txCtx context.Context) error {
				var err error
				perms, roles, dsPtr, orgID, err = permSvc.LoadUserSecurityContext(txCtx, claims.UserID)
				return err
			})
			if err == nil && dsPtr != nil {
				ds = *dsPtr
			}
		}

		return &xinContext.UserContext{
			XinContext:  xc,
			OrgID:       orgID,
			Roles:       roles,
			Permissions: perms,
			DataScope:   ds,
		}
	})

	c.Request = c.Request.WithContext(ctx)
}

// Auth 认证中间件 - 验证 JWT Token 和 Session
func Auth(cfg *config.JWTConfig, sm session.SessionManager, permSvc SecurityContextLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := processAuthToken(c, cfg, sm)
		if err != nil {
			if err == jwt.ErrTokenUnverifiable {
				resp.Unauthorized(c, "unauthorized")
			} else if err == jwt.ErrTokenExpired {
				resp.Unauthorized(c, "session expired or revoked")
			} else {
				resp.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		injectAuthContext(c, claims, permSvc)
		c.Next()
	}
}

// AuthLite 轻量级认证中间件 - 只验证 Token 和注入 XinContext，不加载权限数据
// 适用于只需要知道用户身份但不需要权限检查的场景（如公开接口的个性化内容）
//
// 注意：虽然 UserContext 采用懒加载且只会加载一次，但 AuthLite 有以下优势：
// 1. 明确表达“此路由不需要权限”的意图
// 2. 防止误调用 MustNewUserContext 导致意外加载权限
// 3. 减少内存占用（不注册 UserContextLoader）
// 4. 更安全（从根源上杜绝权限数据被访问的可能）
func AuthLite(cfg *config.JWTConfig, sm session.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := processAuthToken(c, cfg, sm)
		if err != nil {
			if err == jwt.ErrTokenUnverifiable {
				resp.Unauthorized(c, "unauthorized")
			} else if err == jwt.ErrTokenExpired {
				resp.Unauthorized(c, "session expired or revoked")
			} else {
				resp.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// 只注入 XinContext，不注册 UserContextLoader
		ctx := c.Request.Context()
		var xc *xinContext.XinContext
		if existingXc, ok := xinContext.XinContextFrom(ctx); ok {
			xc = existingXc.Clone()
			xc.TenantID = claims.TenantID
			xc.UserID = claims.UserID
			xc.SessionID = claims.SessionID
			xc.Role = claims.Role
			xc.PlatformRoles = claims.PlatformRoles
		} else {
			xc = &xinContext.XinContext{
				TenantID:      claims.TenantID,
				UserID:        claims.UserID,
				SessionID:     claims.SessionID,
				Role:          claims.Role,
				PlatformRoles: claims.PlatformRoles,
			}
		}
		ctx = xinContext.WithXinContext(ctx, xc)
		ctx = xinContext.WithTenantID(ctx, claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件 - 如果有 Token 则解析并注入上下文，没有或无效也继续执行
func OptionalAuth(cfg *config.JWTConfig, sm session.SessionManager, permSvc SecurityContextLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := processAuthToken(c, cfg, sm)
		// 如果 Token 解析成功，注入上下文；否则尝试从 Header 获取租户信息（当游客）
		if err == nil && claims != nil {
			injectAuthContext(c, claims, permSvc)
		} else {
			// 兜底：如果没传 Token（或无效），尝试提取 X-Tenant-ID，以便公共接口也能拿到租户上下文
			if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
				if tenantID, parseErr := strconv.ParseUint(tenantIDStr, 10, 64); parseErr == nil {
					tid := uint(tenantID)
					var xc *xinContext.XinContext
					if existingXc, ok := xinContext.XinContextFrom(c.Request.Context()); ok {
						xc = existingXc.Clone()
						xc.TenantID = tid
					} else {
						xc = &xinContext.XinContext{TenantID: tid}
					}
					c.Request = c.Request.WithContext(xinContext.WithTenantID(xinContext.WithXinContext(c.Request.Context(), xc), tid))
				}
			}
		}
		c.Next()
	}
}

// Require creates an authorization middleware from a typed permission spec.
// Thin wrapper around pkg/middleware.Require to keep a single source of truth.
func Require(spec permission.Spec) gin.HandlerFunc {
	return pkgmiddleware.Require(spec)
}

// RequireAuthenticated creates a login-only middleware with no RBAC check.
func RequireAuthenticated() gin.HandlerFunc {
	return pkgmiddleware.RequireAuthenticated()
}

// RequireAny creates an authorization middleware that accepts any spec.
func RequireAny(specs ...permission.Spec) gin.HandlerFunc {
	return pkgmiddleware.RequireAny(specs...)
}

// RequireAll creates an authorization middleware that requires all specs.
func RequireAll(specs ...permission.Spec) gin.HandlerFunc {
	return pkgmiddleware.RequireAll(specs...)
}

// RequirePlatformRole 校验当前登录账号是否携带指定的平台级角色（如 super_admin）。
//
// 设计意图：跨租户 / 平台级操作（如租户管理、计费管理、平台字典）必须显式校验
// 平台角色，不能仅依赖资源权限码——因为资源权限是租户内的 RBAC，无法表达
// "跨越所有租户"的特权。
//
// 注意：该中间件依赖 Auth 中间件先注入 XinContext.PlatformRoles。
// 使用方式：在 protected 路由分组之后链式追加，或在单条路由上叠加。
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

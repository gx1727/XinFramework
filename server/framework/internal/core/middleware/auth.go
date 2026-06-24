package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/xincontext"
	"gx1727.com/xin/framework/pkg/db"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
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

// injectBaseContext 把 JWT claims 注入到请求上下文里的 XinContext + TenantID。
//
// 行为：
//   - 如果 ctx 上已经有 XinContext，clone 后覆盖身份字段（保留其他字段）；
//   - 否则新建一个 XinContext；
//   - 总是 WithXinContext + WithTenantID 写回 ctx。
//
// 仅注入 XinContext（轻量身份）。如果还需要 UserContext（RBAC + DataScope），
// 调用方应当接着注册 UserContextLoader（见 injectAuthContext）。
func injectBaseContext(c *gin.Context, claims *jwtpkg.Claims) {
	ctx := c.Request.Context()
	var xc *xincontext.XinContext
	if existingXc, ok := xincontext.XinContextFrom(ctx); ok {
		xc = existingXc.Clone()
		xc.TenantID = claims.TenantID
		xc.UserID = claims.UserID
		xc.SessionID = claims.SessionID
		xc.Role = claims.Role
		xc.PlatformRoles = claims.PlatformRoles
	} else {
		xc = &xincontext.XinContext{
			TenantID:      claims.TenantID,
			UserID:        claims.UserID,
			SessionID:     claims.SessionID,
			Role:          claims.Role,
			PlatformRoles: claims.PlatformRoles,
		}
	}
	ctx = xincontext.WithXinContext(ctx, xc)
	ctx = xincontext.WithTenantID(ctx, claims.TenantID)
	c.Request = c.Request.WithContext(ctx)
}

// handleAuthError 把 processAuthToken 抛出的 jwt 错误翻译成统一的 HTTP 401。
// 区分三种情况：
//   - ErrTokenUnverifiable：无 token 或格式错误 → "unauthorized"
//   - ErrTokenExpired：session 失效/被吊销 → "session expired or revoked"
//   - 其他（签名错误、claims 无效等）→ "invalid token"
//
// 调用方需要在 handleAuthError 之后立即 c.Abort() 并 return。
func handleAuthError(c *gin.Context, err error) {
	switch {
	case err == jwt.ErrTokenUnverifiable:
		resp.Unauthorized(c, "unauthorized")
	case err == jwt.ErrTokenExpired:
		resp.Unauthorized(c, "session expired or revoked")
	default:
		resp.Unauthorized(c, "invalid token")
	}
	c.Abort()
}

// injectAuthContext loads permissions and injects UserContext and XinContext into the request
func injectAuthContext(c *gin.Context, claims *jwtpkg.Claims, permSvc SecurityContextLoader, pool *pgxpool.Pool) {
	ctx := c.Request.Context()
	xc, _ := xincontext.XinContextFrom(ctx)
	if xc == nil {
		// 没有 base context 时先补上——理论上前置 injectBaseContext 已写入，
		// 但保留独立路径避免依赖顺序。
		xc = &xincontext.XinContext{
			TenantID:      claims.TenantID,
			UserID:        claims.UserID,
			SessionID:     claims.SessionID,
			Role:          claims.Role,
			PlatformRoles: claims.PlatformRoles,
		}
		ctx = xincontext.WithXinContext(ctx, xc)
		ctx = xincontext.WithTenantID(ctx, claims.TenantID)
	}

	// 注册懒加载生成器到 Context 的某个钩子里，
	// 当实际业务中有人调用 MustNewUserContext 时才去查 DB 构建 UserContext
	ctx = xincontext.WithUserContextLoader(ctx, func() *xincontext.UserContext {
		var roles []string
		var perms map[string]bool
		var ds permission.DataScope
		var orgID int64

		if permSvc != nil {
			var dsPtr *permission.DataScope
			// 因为 RLS 策略强制要求 tenant_id，我们需要包裹在租户事务中
			// 只有设置了 app.tenant_id 才能查询出 users/roles/permissions
			err := db.RunInTenantTx(ctx, pool, claims.TenantID, func(txCtx context.Context) error {
				var err error
				perms, roles, dsPtr, orgID, err = permSvc.LoadUserSecurityContext(txCtx, claims.UserID)
				return err
			})
			if err == nil && dsPtr != nil {
				ds = *dsPtr
			}
		}

		return &xincontext.UserContext{
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
func Auth(cfg *config.JWTConfig, sm session.SessionManager, permSvc SecurityContextLoader, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := processAuthToken(c, cfg, sm)
		if err != nil {
			handleAuthError(c, err)
			return
		}

		injectBaseContext(c, claims)
		injectAuthContext(c, claims, permSvc, pool)
		c.Next()
	}
}

// AuthLite 轻量级认证中间件 - 只验证 Token 和注入 XinContext，不加载权限数据
// 适用于只需要知道用户身份但不需要权限检查的场景（如公开接口的个性化内容）
//
// 注意：虽然 UserContext 采用懒加载且只会加载一次，但 AuthLite 有以下优势：
// 1. 明确表达"此路由不需要权限"的意图
// 2. 防止误调用 MustNewUserContext 导致意外加载权限
// 3. 减少内存占用（不注册 UserContextLoader）
// 4. 更安全（从根源上杜绝权限数据被访问的可能）
func AuthLite(cfg *config.JWTConfig, sm session.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := processAuthToken(c, cfg, sm)
		if err != nil {
			handleAuthError(c, err)
			return
		}

		injectBaseContext(c, claims)
		c.Next()
	}
}

// OptionalAuth 可选认证中间件 - 如果有 Token 则解析并注入上下文，没有或无效也继续执行
func OptionalAuth(cfg *config.JWTConfig, sm session.SessionManager, permSvc SecurityContextLoader, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := processAuthToken(c, cfg, sm)
		// 如果 Token 解析成功，注入上下文；否则尝试从 Header 获取租户信息（当游客）
		if err == nil && claims != nil {
			injectBaseContext(c, claims)
			injectAuthContext(c, claims, permSvc, pool)
		} else {
			// 兜底：如果没传 Token（或无效），尝试提取 X-Tenant-ID，以便公共接口也能拿到租户上下文
			if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
				if tenantID, parseErr := strconv.ParseUint(tenantIDStr, 10, 64); parseErr == nil {
					tid := uint(tenantID)
					var xc *xincontext.XinContext
					if existingXc, ok := xincontext.XinContextFrom(c.Request.Context()); ok {
						xc = existingXc.Clone()
						xc.TenantID = tid
					} else {
						xc = &xincontext.XinContext{TenantID: tid}
					}
					c.Request = c.Request.WithContext(xincontext.WithTenantID(xincontext.WithXinContext(c.Request.Context(), xc), tid))
				}
			}
		}
		c.Next()
	}
}

// Require* and RequirePlatformRole live in pkg/middleware. This file
// used to expose thin wrappers for backward compatibility, but every
// call site now imports pkg/middleware directly (Phase 7 cleanup).

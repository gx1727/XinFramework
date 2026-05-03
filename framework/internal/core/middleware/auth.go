package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/framework/pkg/config"
	xinContext "gx1727.com/xin/framework/pkg/context"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/permission"
	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/framework/pkg/session"
)

// PermissionServiceInterface defines the permission service methods needed by Auth middleware
type PermissionServiceInterface interface {
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
func injectAuthContext(c *gin.Context, claims *jwtpkg.Claims, permSvc PermissionServiceInterface) {
	var roles []string
	var perms map[string]bool
	var ds permission.DataScope
	var orgID int64

	if permSvc != nil {
		ctx := c.Request.Context()
		var dsPtr *permission.DataScope
		perms, roles, dsPtr, orgID, _ = permSvc.LoadUserSecurityContext(ctx, claims.UserID)
		if dsPtr != nil {
			ds = *dsPtr
		}
	}

	ctx := c.Request.Context()

	// Update XinContext
	var xc *xinContext.XinContext
	if existingXc, ok := xinContext.XinContextFrom(ctx); ok {
		xc = existingXc.Clone()
		xc.TenantID = claims.TenantID
		xc.UserID = claims.UserID
		xc.SessionID = claims.SessionID
		xc.Role = claims.Role
	} else {
		xc = &xinContext.XinContext{
			TenantID:  claims.TenantID,
			UserID:    claims.UserID,
			SessionID: claims.SessionID,
			Role:      claims.Role,
		}
	}
	ctx = xinContext.WithXinContext(ctx, xc)

	// Create UserContext
	uc := &xinContext.UserContext{
		XinContext:  xc,
		OrgID:       orgID,
		Roles:       roles,
		Permissions: perms,
		DataScope:   ds,
	}

	ctx = xinContext.WithUserContext(ctx, uc)
	ctx = xinContext.WithTenantID(ctx, claims.TenantID)

	c.Request = c.Request.WithContext(ctx)
}

// Auth 认证中间件 - 验证 JWT Token 和 Session
func Auth(cfg *config.JWTConfig, sm session.SessionManager, permSvc PermissionServiceInterface) gin.HandlerFunc {
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

// OptionalAuth 可选认证中间件 - 如果有 Token 则解析并注入上下文，没有或无效也继续执行
func OptionalAuth(cfg *config.JWTConfig, sm session.SessionManager, permSvc PermissionServiceInterface) gin.HandlerFunc {
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

// RequirePermission 创建权限检查中间件 - 检查特定权限
// 用法: protected.GET("/users", RequirePermission("user", "list"), h.List)
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc := xinContext.MustNewUserContext(c)

		if !uc.HasPermission(resource, action) {
			resp.Forbidden(c, "permission denied: "+resource+":"+action)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 创建权限检查中间件 - 用户拥有任意一个权限即可通过
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc := xinContext.MustNewUserContext(c)

		for _, perm := range permissions {
			parts := strings.SplitN(perm, ":", 2)
			if len(parts) != 2 {
				continue
			}
			if uc.HasPermission(parts[0], parts[1]) {
				c.Next()
				return
			}
		}

		resp.Forbidden(c, "permission denied")
		c.Abort()
	}
}

// RequireAllPermissions 创建权限检查中间件 - 用户必须拥有所有权限才能通过
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc := xinContext.MustNewUserContext(c)

		for _, perm := range permissions {
			parts := strings.SplitN(perm, ":", 2)
			if len(parts) != 2 {
				resp.Forbidden(c, "invalid permission format: "+perm)
				c.Abort()
				return
			}
			if !uc.HasPermission(parts[0], parts[1]) {
				resp.Forbidden(c, "permission denied: "+perm)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

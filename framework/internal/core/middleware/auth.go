package middleware

import (
	"context"
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
	LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error)
	LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error)
	LoadRoles(ctx context.Context, userID uint) ([]string, error)
	GetUserOrgID(ctx context.Context, userID uint) (int64, error)
}

// Auth 认证中间件 - 验证 JWT Token 和 Session
func Auth(cfg *config.JWTConfig, sm session.SessionManager, permSvc PermissionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			resp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &jwtpkg.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			resp.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		ok, err := sm.Validate(claims.SessionID)
		if err != nil || !ok {
			resp.Unauthorized(c, "session expired or revoked")
			c.Abort()
			return
		}

		// Prepare UserContext fields
		var roles []string
		var perms map[string]bool
		var ds permission.DataScope
		var orgID int64

		if permSvc != nil {
			ctx := c.Request.Context()
			perms, _ = permSvc.LoadPermissions(ctx, claims.UserID)
			roles, _ = permSvc.LoadRoles(ctx, claims.UserID)
			dsPtr, _ := permSvc.LoadDataScope(ctx, claims.UserID)
			if dsPtr != nil {
				ds = *dsPtr
			}
			orgID, _ = permSvc.GetUserOrgID(ctx, claims.UserID)
		}

		// Create UserContext
		uc := &xinContext.UserContext{
			TenantID:    claims.TenantID,
			UserID:      claims.UserID,
			OrgID:       orgID,
			SessionID:   claims.SessionID,
			Roles:       roles,
			Permissions: perms,
			DataScope:   ds,
		}

		ctx := c.Request.Context()
		ctx = xinContext.WithUserContext(ctx, uc)
		ctx = xinContext.WithTenantID(ctx, claims.TenantID)

		// Also update XinContext if present
		if xc, ok := xinContext.XinContextFrom(ctx); ok {
			xc.SetTenantID(claims.TenantID)
			xc.SetUserID(claims.UserID)
			xc.SetSessionID(claims.SessionID)
			xc.SetRole(claims.Role)
		} else {
			xc = &xinContext.XinContext{
				TenantID:  claims.TenantID,
				UserID:    claims.UserID,
				SessionID: claims.SessionID,
				Role:      claims.Role,
			}
			ctx = xinContext.WithXinContext(ctx, xc)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("session_id", claims.SessionID)
		c.Set("role", claims.Role)
		c.Set("roles", roles)

		c.Next()
	}
}

// RequirePermission 创建权限检查中间件 - 检查特定权限
// 用法: protected.GET("/users", RequirePermission("user", "list"), h.List)
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc := xinContext.NewUserContext(c)
		if uc.UserID == 0 {
			resp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

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
		uc := xinContext.NewUserContext(c)
		if uc.UserID == 0 {
			resp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

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
		uc := xinContext.NewUserContext(c)
		if uc.UserID == 0 {
			resp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

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

package middleware

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gx1727.com/xin/framework/pkg/config"
	xinContext "gx1727.com/xin/framework/pkg/context"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/logger"
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

func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	if cfg == nil || !cfg.Enabled || len(cfg.AllowOrigins) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowOrigin := ""
		for _, o := range cfg.AllowOrigins {
			o = strings.TrimSpace(o)
			if o == "*" {
				allowOrigin = "*"
				break
			}
			if strings.EqualFold(o, origin) {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin == "" {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)
		c.Header("Access-Control-Allow-Headers", cfg.AllowHeaders)
		c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

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

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func Tenant(mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mode == "" {
			c.Next()
			return
		}

		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			if tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64); err == nil {
				tid := uint(tenantID)
				ctx := xinContext.New(c)
				ctx.SetTenantID(tid)
				c.Request = c.Request.WithContext(xinContext.WithTenantID(xinContext.WithXinContext(c.Request.Context(), ctx), tid))
				c.Set("tenant_id", tid)
			}
		}

		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		requestID, _ := c.Get("request_id")
		reqID, _ := requestID.(string)
		if reqID == "" {
			reqID = "-"
		}

		if raw != "" {
			path = path + "?" + raw
		}

		switch {
		case statusCode >= 500:
			logger.Errorf("[%s] %s %s | %d | %v | %s", reqID, method, path, statusCode, latency, clientIP)
		case statusCode >= 400:
			logger.Warnf("[%s] %s %s | %d | %v | %s", reqID, method, path, statusCode, latency, clientIP)
		default:
			logger.Infof("[%s] %s %s | %d | %v | %s", reqID, method, path, statusCode, latency, clientIP)
		}
	}
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// RequirePermission creates middleware that checks for a specific permission
// Usage: protected.GET("/users", RequirePermission("user", "list"), h.List)
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

// RequireAnyPermission creates middleware that passes if user has ANY of the permissions
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

// RequireAllPermissions creates middleware that passes only if user has ALL permissions
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

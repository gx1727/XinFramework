// Package jwt 提供 JWT token 的签发、解析与 Claims 类型。
//
// 本包仅负责“令牌本身”的生成与校验，会话有效性由 session.SessionManager
// 独立检查。Auth 中间件在中间件侧会把这两步串联起来。
//
// 关键概念：
//   - TokenTypeAccess / TokenTypeRefresh：区分访问令牌与刷新令牌
//   - PlatformRoleSuperAdmin：平台超级管理员角色编码，跨租户特权
//   - Claims.PlatformRoles：登录时查询 sys_user_roles → sys_roles.code 填入，
//     中间件通过 jwt.PlatformRoleSuperAdmin 常量识别并短路放行
//
// 使用流程：
//  1. 登录成功 → GenerateWithPlatformRoles 签发 access + refresh 两个 token
//  2. 客户端调用 API 时带 Authorization: Bearer <access>
//  3. Auth 中间件调 Validate 解析 + session.Manager().Validate(SessionID) 检查存活
//  4. 接近过期时调 refresh 接口重签发
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/framework/pkg/config"
)

// Token 类型常量。
const (
	TokenTypeAccess  = "access"  // 访问令牌
	TokenTypeRefresh = "refresh" // 刷新令牌
)

// PlatformRoleSuperAdmin 平台超级管理员角色编码。
// 拥有此角色的账号可以无视租户边界与 DataScope，中间件 requireWithSpecs 会
// 检到 PlatformRoles 含本常量后直接放行。
const PlatformRoleSuperAdmin = "super_admin"

// Claims 是框架自定义的 JWT 载荷。
//
// UserID / TenantID / Role 在 Auth 中间件被取出注入到 XinContext；
// SessionID 供 session.Manager().Validate() 做存活检查；
// PlatformRoles 供 requireWithSpecs 判断 super_admin 短路。
type Claims struct {
	UserID        uint     `json:"user_id"`
	TenantID      uint     `json:"tenant_id"`
	Role          string   `json:"role"`
	SessionID     string   `json:"sid"`
	TokenType     string   `json:"token_type"`
	PlatformRoles []string `json:"platform_roles,omitempty"`
	jwt.RegisteredClaims
}

// Generate 签发默认的访问令牌（TokenTypeAccess）。
func Generate(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string) (string, error) {
	return GenerateWithType(cfg, userID, tenantID, role, sessionID, TokenTypeAccess)
}

// GenerateWithType 签发指定类型的令牌（access 或 refresh）。
func GenerateWithType(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, tokenType string) (string, error) {
	return GenerateWithPlatformRoles(cfg, userID, tenantID, role, sessionID, nil, tokenType)
}

// GenerateWithPlatformRoles 签发携带平台角色信息的令牌。
// platformRoles 可为 nil；非空时 JWT 中会带上这些角色码，
// 中间件据此识别 super_admin 并短路放行。
func GenerateWithPlatformRoles(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, platformRoles []string, tokenType string) (string, error) {
	expire := cfg.Expire
	if tokenType == TokenTypeRefresh {
		expire = cfg.RefreshExpire
	}

	claims := Claims{
		UserID:        userID,
		TenantID:      tenantID,
		Role:          role,
		SessionID:     sessionID,
		TokenType:     tokenType,
		PlatformRoles: platformRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// HasPlatformRole 便捷判断：claims 是否包含指定的平台级角色。
// 典型调用：HasPlatformRole(PlatformRoleSuperAdmin)。
func (c *Claims) HasPlatformRole(role string) bool {
	if c == nil || role == "" {
		return false
	}
	for _, r := range c.PlatformRoles {
		if r == role {
			return true
		}
	}
	return false
}

func Validate(tokenString string, cfg *config.JWTConfig) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ValidateRefresh(tokenString string, cfg *config.JWTConfig) (*Claims, error) {
	claims, err := Validate(tokenString, cfg)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}
	return claims, nil
}

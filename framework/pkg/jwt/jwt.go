package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/framework/pkg/config"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// PlatformRoleSuperAdmin 平台级超级管理员角色编码
// 拥有此角色的账号可以无视租户/数据范围限制
const PlatformRoleSuperAdmin = "super_admin"

type Claims struct {
	UserID        uint     `json:"user_id"`
	TenantID      uint     `json:"tenant_id"`
	Role          string   `json:"role"`
	SessionID     string   `json:"sid"`
	TokenType     string   `json:"token_type"`
	PlatformRoles []string `json:"platform_roles,omitempty"`
	jwt.RegisteredClaims
}

func Generate(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string) (string, error) {
	return GenerateWithType(cfg, userID, tenantID, role, sessionID, TokenTypeAccess)
}

func GenerateWithType(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, tokenType string) (string, error) {
	return GenerateWithPlatformRoles(cfg, userID, tenantID, role, sessionID, nil, tokenType)
}

// GenerateWithPlatformRoles 签发带平台角色信息的 Token。
// platformRoles 可为空；为非空时 JWT 中会携带这些角色，中间件据此识别 super_admin。
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

// IsSuperAdmin 便捷判断：claims 是否包含 super_admin 平台角色
func (c *Claims) IsSuperAdmin() bool {
	if c == nil {
		return false
	}
	for _, r := range c.PlatformRoles {
		if r == PlatformRoleSuperAdmin {
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

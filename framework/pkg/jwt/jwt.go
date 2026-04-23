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

type Claims struct {
	UserID    uint   `json:"user_id"`
	TenantID  uint   `json:"tenant_id"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func Generate(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string) (string, error) {
	return GenerateWithType(cfg, userID, tenantID, role, sessionID, TokenTypeAccess)
}

func GenerateWithType(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, tokenType string) (string, error) {
	expire := cfg.Expire
	if tokenType == TokenTypeRefresh {
		expire = cfg.RefreshExpire
	}

	claims := Claims{
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		SessionID: sessionID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
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

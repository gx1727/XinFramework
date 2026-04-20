package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/pkg/config"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	TenantID uint   `json:"tenant_id"`
	Role     string `json:"role"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func Generate(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string) (string, error) {
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.Expire) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

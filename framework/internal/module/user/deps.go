package user

import (
	"time"

	"gorm.io/gorm"
	"gx1727.com/xin/framework/pkg/config"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/session"
)

type Dependencies struct {
	DB      *gorm.DB
	Config  *config.Config
	Session SessionManager
}

type defaultSessionManager struct{}

func (defaultSessionManager) Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	return session.Create(sessionID, userID, tenantID, role, ttl)
}

func (defaultSessionManager) Revoke(sessionID string) error {
	return session.Revoke(sessionID)
}

func DefaultDependencies(cfg *config.Config, db *gorm.DB) Dependencies {
	return Dependencies{
		DB:      db,
		Config:  cfg,
		Session: defaultSessionManager{},
	}
}

func (d Dependencies) jwtConfig() *config.JWTConfig {
	if d.Config == nil {
		return nil
	}
	return &d.Config.JWT
}

func (d Dependencies) generateToken(userID, tenantID uint, role, sessionID string) (string, error) {
	jwtCfg := d.jwtConfig()
	if jwtCfg == nil {
		return "", ErrBackendUnavailable
	}
	return jwtpkg.Generate(jwtCfg, userID, tenantID, role, sessionID)
}

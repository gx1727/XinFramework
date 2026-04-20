package auth

import (
	"gx1727.com/xin/internal/infra/logger"
	"gx1727.com/xin/pkg/config"
)

type AuthConfig struct {
	MaxLoginAttempts         int    `yaml:"max_login_attempts"`
	LockDurationSec          int    `yaml:"lock_duration_sec"`
	PasswordPolicy           string `yaml:"password_policy"`
	TokenExpireSec           int    `yaml:"token_expire_sec"`
	RefreshTokenExpireSec    int    `yaml:"refresh_token_expire_sec"`
}

var moduleCfg *AuthConfig

func Cfg() *AuthConfig {
	return moduleCfg
}

func InitConfig() error {
	moduleCfg = &AuthConfig{
		MaxLoginAttempts:      5,
		LockDurationSec:       300,
		PasswordPolicy:        "standard",
		TokenExpireSec:        3600,
		RefreshTokenExpireSec: 86400,
	}
	if err := config.LoadModule("auth", moduleCfg); err != nil {
		return err
	}
	logger.Infof("auth module config loaded: attempts=%d lock=%ds policy=%s",
		moduleCfg.MaxLoginAttempts, moduleCfg.LockDurationSec, moduleCfg.PasswordPolicy)
	return nil
}

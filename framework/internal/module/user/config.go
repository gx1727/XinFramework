package user

import (
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/logger"
)

type AuthConfig struct {
	MaxLoginAttempts      int    `yaml:"max_login_attempts"`
	LockDurationSec       int    `yaml:"lock_duration_sec"`
	PasswordPolicy        string `yaml:"password_policy"`
	TokenExpireSec        int    `yaml:"token_expire_sec"`
	RefreshTokenExpireSec int    `yaml:"refresh_token_expire_sec"`
}

var authCfg *AuthConfig
var userLogger *logger.Logger

func Cfg() *AuthConfig {
	return authCfg
}

func InitConfig() error {
	userLogger = logger.Module("user")

	authCfg = &AuthConfig{
		MaxLoginAttempts:      5,
		LockDurationSec:       300,
		PasswordPolicy:        "standard",
		TokenExpireSec:        3600,
		RefreshTokenExpireSec: 86400,
	}

	if err := config.LoadModule("user", authCfg); err != nil {
		return err
	}

	if userLogger != nil {
		userLogger.Infof("user module config loaded: attempts=%d lock=%ds policy=%s",
			authCfg.MaxLoginAttempts, authCfg.LockDurationSec, authCfg.PasswordPolicy)
	} else {
		logger.Infof("user module config loaded: attempts=%d lock=%ds policy=%s",
			authCfg.MaxLoginAttempts, authCfg.LockDurationSec, authCfg.PasswordPolicy)
	}
	return nil
}

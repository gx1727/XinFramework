package auth

import (
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/logger"
)

var authLogger *logger.Logger

func Cfg() *config.AuthConfig {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}
	return &cfg.Auth
}

func InitConfig() error {
	authLogger = logger.Module("auth")

	cfg := config.Get()
	if cfg != nil && cfg.Auth.MaxLoginAttempts == 0 {
		cfg.Auth.MaxLoginAttempts = 5
		cfg.Auth.LockDurationSec = 300
		cfg.Auth.PasswordPolicy = "standard"
		cfg.Auth.TokenExpireSec = 3600
		cfg.Auth.RefreshTokenExpireSec = 86400
	}

	if authLogger != nil {
		authLogger.Infof("auth module config loaded: attempts=%d lock=%ds policy=%s",
			cfg.Auth.MaxLoginAttempts, cfg.Auth.LockDurationSec, cfg.Auth.PasswordPolicy)
	} else {
		logger.Infof("auth module config loaded: attempts=%d lock=%ds policy=%s",
			cfg.Auth.MaxLoginAttempts, cfg.Auth.LockDurationSec, cfg.Auth.PasswordPolicy)
	}
	return nil
}

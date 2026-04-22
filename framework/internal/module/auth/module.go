package auth

import (
	"gx1727.com/xin/framework/pkg/logger"
)

func InitConfig() error {
	l := logger.Module("auth")
	if l != nil {
		l.Infof("auth module loaded (role/permission not yet implemented)")
	} else {
		logger.Infof("auth module loaded (role/permission not yet implemented)")
	}
	return nil
}

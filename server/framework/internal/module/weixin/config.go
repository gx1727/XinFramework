package weixin

import (
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/logger"
)

type WxxcxConfig struct {
	AppID     string `yaml:"appid"`
	AppSecret string `yaml:"appsecret"`
}

var wxxcxCfg *WxxcxConfig
var weixinLogger *logger.Logger

func Cfg() *WxxcxConfig {
	return wxxcxCfg
}

func InitConfig() error {
	weixinLogger = logger.Module("weixin")

	wxxcxCfg = &WxxcxConfig{}

	if err := config.LoadModule("weixin", wxxcxCfg); err != nil {
		return err
	}

	if wxxcxCfg.AppID == "" || wxxcxCfg.AppSecret == "" {
		if weixinLogger != nil {
			weixinLogger.Warnf("weixin config not loaded: appid=%q appsecret=%q", wxxcxCfg.AppID, wxxcxCfg.AppSecret)
		}
	} else {
		if weixinLogger != nil {
			weixinLogger.Infof("weixin module config loaded: appid=%s", maskAppID(wxxcxCfg.AppID))
		} else {
			logger.Infof("weixin module config loaded: appid=%s", maskAppID(wxxcxCfg.AppID))
		}
	}
	return nil
}

func maskAppID(appid string) string {
	if len(appid) <= 8 {
		return "***"
	}
	return appid[:4] + "..." + appid[len(appid)-4:]
}

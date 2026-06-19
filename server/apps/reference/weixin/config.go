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

// weixinCfg 持有整个 *config.Config，主要给 generateTokens 用 JWT 配置。
// 与 wxxcxCfg 的区别：wxxcxCfg 只装 weixin 自己的子段，weixinCfg 是全局。
// 过渡期字段，未来 main.go 显式 Build 后应改为显式注入到 NewService。
var weixinCfg *config.Config

func Cfg() *WxxcxConfig {
	return wxxcxCfg
}

// SetGlobalConfig 让模块入口在 boot 阶段把全局 config 注入进来，
// 避免 service 直接调用 pkgconfig.Get()。
func SetGlobalConfig(c *config.Config) {
	weixinCfg = c
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

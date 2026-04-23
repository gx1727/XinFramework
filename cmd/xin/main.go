package main

import (
	"log"

	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/module/cms"
)

// moduleRegistry 存储所有可用的外部插件
// 添加新 app 只需在这里添加一行
var moduleRegistry = map[string]func() plugin.Module{
	"cms": func() plugin.Module { return cms.Module(cms.NewHandler()) },
	// future apps: "shop": shop.Module, "blog": blog.Module, ...
}

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// 根据 config.yaml 中的 apps 配置动态注册模块
	for _, app := range cfg.Apps {
		if factory, ok := moduleRegistry[app]; ok {
			framework.RegisterModule(factory())
		}
	}

	framework.Run(cfg)
}

package main

import (
	"log"

	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/module/cms"
)

var moduleRegistry = map[string]func() plugin.Module{
	"cms": cms.Module,
}

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	for _, app := range cfg.Apps {
		if factory, ok := moduleRegistry[app]; ok {
			framework.RegisterModule(factory())
		}
	}

	framework.Run(cfg)
}

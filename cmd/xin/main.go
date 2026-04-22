package main

import (
	"log"

	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/module/cms"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if cfg.AppEnabled("cms") {
		framework.RegisterModule(cms.Module())
	}

	framework.Run(cfg)
}

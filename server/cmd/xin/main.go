package main

import (
	"log"

	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"

	// Phase 2: boot modules
	_ "gx1727.com/xin/apps/boot/auth"
	_ "gx1727.com/xin/apps/boot/tenant"

	// Phase 3: RBAC suite
	_ "gx1727.com/xin/apps/rbac/menu"
	_ "gx1727.com/xin/apps/rbac/organization"
	_ "gx1727.com/xin/apps/rbac/permission"
	_ "gx1727.com/xin/apps/rbac/resource"
	_ "gx1727.com/xin/apps/rbac/role"
	_ "gx1727.com/xin/apps/rbac/user"

	// Optional apps
	_ "gx1727.com/xin/apps/reference/asset"
	_ "gx1727.com/xin/apps/reference/dict"
	_ "gx1727.com/xin/apps/reference/weixin"
	_ "gx1727.com/xin/apps/reference/config"
	_ "gx1727.com/xin/apps/system"

	// External apps
	_ "gx1727.com/xin/apps/cms"
	_ "gx1727.com/xin/apps/flag"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// All modules — including the ones that previously lived in
	// framework/builtin_modules.go before Phase 2 — register themselves via
	// the side-effect imports above. By the time main() reaches this point,
	// plugin.Apps() contains every module; framework.Run reads cfg.Module
	// as the enable-list and calls Init / Register on each enabled one.
	framework.Run(cfg)
}

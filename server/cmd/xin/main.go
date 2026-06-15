package main

import (
	"log"

	"gx1727.com/xin/framework"
	"gx1727.com/xin/framework/pkg/config"

	// External apps + Phase 2 boot modules. Each module's init()
	// registers itself through plugin.Register. main.go no longer
	// maintains a hardcoded map of every module.
	//
	// Phase 2 status: auth + tenant have moved to apps/boot/.
	// Phase 3 will move RBAC (user/role/menu/...) to apps/rbac/.
	_ "gx1727.com/xin/apps/boot/auth"
	_ "gx1727.com/xin/apps/boot/tenant"

	_ "gx1727.com/xin/apps/cms"
	_ "gx1727.com/xin/apps/flag"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// Remaining framework-internal modules are pulled in transitively
	// through "gx1727.com/xin/framework" (see framework/builtin_modules.go).
	// After framework.Run completes, all built-in + external modules
	// are sitting in plugin.Apps() ready to be initialized.
	framework.Run(cfg)
}
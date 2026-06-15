// Package framework re-exports the built-in modules so that
// cmd/xin/main.go can pull them all in with a single import.
//
// This file exists because of Go's internal/ rule:
// cmd/xin lives in a different module and cannot import from
// framework/internal/. By listing the built-in modules here we
// keep the import inside the framework module, where internal/
// access is allowed.
//
// Each side-effect import triggers the corresponding module's
// init() — which calls plugin.Register(Module()) — so by the time
// framework.Run() executes, all built-in modules are already
// registered through the same plugin.Apps() list as external apps.
//
// Refactor roadmap:
//   Phase 2: auth + tenant have moved out to apps/boot/.
//   Phase 3: RBAC (user/role/menu/resource/permission/organization)
//            will move to apps/rbac/.
//   Phase 3: dict/asset/weixin will move to apps/reference/.
//   When Phase 3 is done this file should be empty and removable.
package framework

import (
	// Side-effect imports: each module's init() registers itself.
	_ "gx1727.com/xin/framework/internal/module/asset"
	_ "gx1727.com/xin/framework/internal/module/dict"
	_ "gx1727.com/xin/framework/internal/module/menu"
	_ "gx1727.com/xin/framework/internal/module/organization"
	_ "gx1727.com/xin/framework/internal/module/permission"
	_ "gx1727.com/xin/framework/internal/module/resource"
	_ "gx1727.com/xin/framework/internal/module/role"
	_ "gx1727.com/xin/framework/internal/module/system"
	_ "gx1727.com/xin/framework/internal/module/user"
	_ "gx1727.com/xin/framework/internal/module/weixin"

	// Phase 2: auth + tenant moved to apps/boot/. They are imported
	// through cmd/xin/main.go via the apps module, NOT here.
)
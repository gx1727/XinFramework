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
//   Phase 1: ✅ unified registration (built-in + external via plugin.Apps())
//   Phase 2: ✅ auth + tenant moved to apps/boot/
//   Phase 3: ✅ RBAC (user/role/menu/resource/permission/organization)
//              moved to apps/rbac/
//   Phase 4: ✅ dict/asset/weixin still in framework — Phase 3b will
//              move them to apps/reference/
//
// Remaining framework/internal/module/ entries are still required
// because weixin (still framework-internal) depends on them transitively
// at startup, and Phase 3b has not yet moved weixin out.
package framework

import (
	// Built-in modules still living in framework/internal/module.
	// Each module's init() registers itself through plugin.Register.
	_ "gx1727.com/xin/framework/internal/module/dict"
	_ "gx1727.com/xin/framework/internal/module/system"
	_ "gx1727.com/xin/framework/internal/module/weixin"

	// Phase 3: auth, tenant, RBAC, asset have all moved to apps/.
	// Imported via cmd/xin/main.go side-effect:
	//   apps/boot/{auth,tenant}      — framework startup required
	//   apps/rbac/{user,role,...}    — RBAC suite
	//   apps/reference/asset         — file storage
)
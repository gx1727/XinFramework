package user

import (
	"context"

	bootauth "gx1727.com/xin/apps/boot/auth"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/db"
)

// newAccountAdapter returns an auth.AccountRepository backed by
// apps/boot/auth's PostgresAccountRepository.
//
// Phase 3: user now lives in apps/rbac/, so it can import
// apps/boot/auth directly. The Phase 2 framework-side adapter
// (which routed through pkgauth.Get()) is no longer needed —
// we just construct the auth-side repository and return it.
//
// If apps/boot/auth's AccountRepository type ever stops satisfying
// pkgauth.AccountRepository (e.g. field shape changes), this file
// will fail to compile and the divergence will be caught at build time.
func newAccountAdapter() pkgauth.AccountRepository {
	return bootauth.NewAccountRepository(db.Get())
}

// Compile-time guard: any change to apps/boot/auth that breaks the
// contract is caught immediately rather than at runtime.
var _ pkgauth.AccountRepository = bootauth.NewAccountRepository(db.Get())

// Suppress unused-import warnings for "context" if all uses are
// removed in future refactors.
var _ = context.TODO
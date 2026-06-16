// Package authz exposes the global AuthorizationService used by
// middleware (Auth, Require, RequireAll) and by business modules
// that need to invalidate the permission cache.
//
// Phase 3 rationale: apps/rbac/{permission,role,resource}/service.go
// need to call GlobalAuthorizationService() at request time. The
// concrete AuthorizationService type lives in framework/internal/service
// (Go's internal/ rule blocks apps/ from importing it). This pkg
// exposes the global accessor with a thin type-erased interface —
// apps use the interface, the framework wires the concrete impl
// in boot.Init.
package authz

import "context"

// Authorization is the public surface apps can consume.
//
// Method signatures mirror what framework/internal/service.AuthorizationService
// exposes. The boot code wires an *AuthorizationService to this
// interface; apps call authz.Get() to retrieve it.
//
// If a new method is added to AuthorizationService that apps need,
// add it here as well.
type Authorization interface {
	// LoadPermissions returns the user's effective permission map
	// (resource_code -> bool).
	LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error)

	// LoadRoles returns the role codes assigned to the user.
	LoadRoles(ctx context.Context, userID uint) ([]string, error)

	// LoadDataScope returns the user's data scope. Returned as
// interface{} — the concrete type is *permission.DataScope.
// Callers that need the concrete type do an `if ds, ok := result.(*permission.DataScope); ...`.
//
// Apps that don't need the concrete type can ignore the value.
	LoadDataScope(ctx context.Context, userID uint) (interface{}, error)

	// InvalidateUser clears cached permissions / data scope for the user.
	InvalidateUser(ctx context.Context, userID uint) error

	// InvalidateRole clears cached data for all users that hold this role.
	InvalidateRole(ctx context.Context, roleID uint) error

	// InvalidateResource clears cached data for all users affected by
	// a resource change.
	InvalidateResource(ctx context.Context, resourceID uint) error
}

var global Authorization

// Set wires the global Authorization. Called from boot.Init.
func Set(a Authorization) {
	global = a
}

// Get returns the global Authorization, or nil if not loaded.
func Get() Authorization {
	return global
}
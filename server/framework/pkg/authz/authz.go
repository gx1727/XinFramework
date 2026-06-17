// Package authz exposes the Authorization interface used by middleware
// (Auth, Require, RequireAll) and by business modules that need to
// invalidate the permission cache.
//
// The concrete *service.AuthorizationService lives in
// framework/internal/service (Go's internal/ rule blocks apps/ from
// importing it). This pkg exposes a thin type-erased interface plus
// a Wrap() adapter so the framework can hand apps an Authorization
// without leaking internal types.
//
// Wiring: boot.Init constructs the concrete service, wraps it via
// Wrap(), and publishes the resulting Authorization onto AppContext
// via appCtx.SetAuthz(...). Apps consume it via ctx.Authz() in their
// module's Register phase.
package authz

import "context"

// Authorization is the public surface apps can consume.
//
// Method signatures mirror what framework/internal/service.AuthorizationService
// exposes. The boot code wires an *AuthorizationService to this
// interface and publishes it via AppContext.SetAuthz.
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
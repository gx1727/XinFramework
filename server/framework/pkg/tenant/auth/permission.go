package auth

import "context"

// RoleResource binds a role to a resource code. Defined here for
// documentation; the concrete type alias lives in apps/tenant/permission/.
type RoleResource struct {
	RoleID       uint
	ResourceID   uint
	ResourceCode string
}

// RoleResourceRepository is the cross-module role-resource binding
// access contract. The concrete implementation in apps/tenant/permission/
// satisfies this interface.
type RoleResourceRepository interface {
	// GetByRoleID returns the resource IDs bound to the given role.
	GetByRoleID(ctx context.Context, roleID uint) ([]uint, error)
	// SetForRole replaces all resource bindings for a role.
	SetForRole(ctx context.Context, roleID uint, resourceIDs []uint) error
	// DeleteByRoleID removes all resource bindings for a role.
	DeleteByRoleID(ctx context.Context, roleID uint) error
}
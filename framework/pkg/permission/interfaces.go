package permission

import (
	"context"
)

// PermissionRepository loads permissions from database
type PermissionRepository interface {
	// GetUserPermissions returns a map of "resource:action" -> true
	GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error)

	// GetUserRoles returns role codes for a user
	GetUserRoles(ctx context.Context, userID uint) ([]string, error)
}

// DataScopeRepository loads data scope from database
type DataScopeRepository interface {
	// GetDataScope returns the data scope for a user
	GetDataScope(ctx context.Context, userID uint) (*DataScope, error)

	// GetUserOrgID returns the user's organization ID
	GetUserOrgID(ctx context.Context, userID uint) (int64, error)

	// GetByRoleID returns org_ids for a role's custom data scope
	GetByRoleID(ctx context.Context, roleID uint) ([]uint, error)

	// SetForRole replaces all data scopes for a role
	SetForRole(ctx context.Context, roleID uint, orgIDs []uint) error
}

// PermissionCache defines caching operations for permissions
type PermissionCache interface {
	GetPermissions(ctx context.Context, userID uint) (map[string]bool, error)
	SetPermissions(ctx context.Context, userID uint, perms map[string]bool) error
	InvalidatePermissions(ctx context.Context, userID uint) error

	GetDataScope(ctx context.Context, userID uint) (*DataScope, error)
	SetDataScope(ctx context.Context, userID uint, ds *DataScope) error
	InvalidateDataScope(ctx context.Context, userID uint) error
}

﻿package permission

import (
	"context"
)

type PermissionRepository interface {
	GetByRoleID(ctx context.Context, roleID uint) ([]Permission, error)
	DeleteByRoleID(ctx context.Context, roleID uint) error
	Create(ctx context.Context, tenantID, roleID uint, p Permission) error
}

type UserPermissionRepository interface {
	GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error)
	GetUserRoles(ctx context.Context, userID uint) ([]string, error)
	GetUserIDsByRole(ctx context.Context, roleID uint) ([]uint, error)
	GetUserIDsByResource(ctx context.Context, resourceID uint) ([]uint, error)
}

type DataScopeRepository interface {
	GetDataScope(ctx context.Context, userID uint) (*DataScope, error)
	GetUserOrgID(ctx context.Context, userID uint) (int64, error)
	GetByRoleID(ctx context.Context, roleID uint) ([]uint, error)
	SetForRole(ctx context.Context, roleID uint, orgIDs []uint) error
}

type PermissionCache interface {
	GetPermissions(ctx context.Context, userID uint) (map[string]bool, error)
	SetPermissions(ctx context.Context, userID uint, perms map[string]bool) error
	InvalidatePermissions(ctx context.Context, userID uint) error

	GetDataScope(ctx context.Context, userID uint) (*DataScope, error)
	SetDataScope(ctx context.Context, userID uint, ds *DataScope) error
	InvalidateDataScope(ctx context.Context, userID uint) error
}

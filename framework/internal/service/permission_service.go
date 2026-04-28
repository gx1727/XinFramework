package service

import (
	"context"
	"fmt"

	"gx1727.com/xin/framework/pkg/permission"
)

// PermissionService handles permission loading and checking
type PermissionService struct {
	permRepo permission.UserPermissionRepository
	dsRepo   permission.DataScopeRepository
	cache    permission.PermissionCache
}

func NewPermissionService(
	permRepo permission.UserPermissionRepository,
	dsRepo permission.DataScopeRepository,
	cache permission.PermissionCache,
) *PermissionService {
	return &PermissionService{
		permRepo: permRepo,
		dsRepo:   dsRepo,
		cache:    cache,
	}
}

// LoadPermissions loads user permissions with caching
func (s *PermissionService) LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	// Try cache first
	if s.cache != nil {
		perms, err := s.cache.GetPermissions(ctx, userID)
		if err == nil && perms != nil {
			return perms, nil
		}
	}

	// Load from database
	perms, err := s.permRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		_ = s.cache.SetPermissions(ctx, userID, perms)
	}

	return perms, nil
}

// LoadDataScope loads user data scope with caching
func (s *PermissionService) LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error) {
	// Try cache first
	if s.cache != nil {
		ds, err := s.cache.GetDataScope(ctx, userID)
		if err == nil && ds != nil {
			return ds, nil
		}
	}

	// Load from database
	ds, err := s.dsRepo.GetDataScope(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load data scope: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		_ = s.cache.SetDataScope(ctx, userID, ds)
	}

	return ds, nil
}

// LoadRoles loads user role codes
func (s *PermissionService) LoadRoles(ctx context.Context, userID uint) ([]string, error) {
	return s.permRepo.GetUserRoles(ctx, userID)
}

// InvalidateUser invalidates all cached permission/data for a user
func (s *PermissionService) InvalidateUser(ctx context.Context, userID uint) error {
	if s.cache != nil {
		_ = s.cache.InvalidatePermissions(ctx, userID)
		_ = s.cache.InvalidateDataScope(ctx, userID)
	}
	return nil
}

// HasPermission checks if user has a specific permission
func (s *PermissionService) HasPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	perms, err := s.LoadPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	return permission.HasPermission(perms, resource, action), nil
}

// BuildDataScopeSQL builds SQL WHERE clause for data filtering
// Returns: "org_id = ANY($1)" or "creator_id = $1", and args slice
func (s *PermissionService) BuildDataScopeSQL(ctx context.Context, userID uint) (string, []any, error) {
	ds, err := s.LoadDataScope(ctx, userID)
	if err != nil {
		return "", nil, err
	}

	orgID, err := s.dsRepo.GetUserOrgID(ctx, userID)
	if err != nil {
		return "", nil, err
	}

	switch ds.Type {
	case permission.DataScopeAll:
		// No filtering - can see all data
		return "", nil, nil

	case permission.DataScopeSelf:
		return "creator_id = $1", []any{userID}, nil

	case permission.DataScopeCustom:
		if len(ds.OrgIDs) == 0 {
			return "creator_id = $1", []any{userID}, nil
		}
		return "org_id = ANY($1)", []any{ds.OrgIDs}, nil

	case permission.DataScopeDept:
		if orgID == 0 {
			return "creator_id = $1", []any{userID}, nil
		}
		return "org_id = $1", []any{orgID}, nil

	case permission.DataScopeDeptAndBelow:
		if orgID == 0 {
			return "creator_id = $1", []any{userID}, nil
		}
		// Use CTE to find all descendant org IDs
		return `
			org_id = $1
			OR org_id IN (
				WITH RECURSIVE org_tree AS (
					SELECT id FROM organizations WHERE id = $1
					UNION ALL
					SELECT o.id FROM organizations o
					JOIN org_tree ot ON o.parent_id = ot.id
				)
				SELECT id FROM org_tree
			)
		`, []any{orgID}, nil

	default:
		return "creator_id = $1", []any{userID}, nil
	}
}

// GetUserOrgID returns the user's organization ID
func (s *PermissionService) GetUserOrgID(ctx context.Context, userID uint) (int64, error) {
	return s.dsRepo.GetUserOrgID(ctx, userID)
}

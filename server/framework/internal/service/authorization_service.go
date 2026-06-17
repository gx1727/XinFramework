package service

import (
	"context"

	"gx1727.com/xin/framework/pkg/permission"
)

// AuthorizationService is the unified facade for RBAC checks, data-scope building,
// and cache invalidation. It wraps the existing PermissionService to keep migration incremental.
type AuthorizationService struct {
	perm *PermissionService
}

func NewAuthorizationService(perm *PermissionService) *AuthorizationService {
	return &AuthorizationService{perm: perm}
}

func (s *AuthorizationService) PermissionService() *PermissionService {
	if s == nil {
		return nil
	}
	return s.perm
}

func (s *AuthorizationService) LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	return s.perm.LoadPermissions(ctx, userID)
}

func (s *AuthorizationService) LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error) {
	return s.perm.LoadDataScope(ctx, userID)
}

func (s *AuthorizationService) LoadRoles(ctx context.Context, userID uint) ([]string, error) {
	return s.perm.LoadRoles(ctx, userID)
}

func (s *AuthorizationService) LoadUserSecurityContext(ctx context.Context, userID uint) (map[string]bool, []string, *permission.DataScope, int64, error) {
	return s.perm.LoadUserSecurityContext(ctx, userID)
}

func (s *AuthorizationService) Can(ctx context.Context, userID uint, spec permission.Spec) (bool, error) {
	if spec.IsAuthOnly() {
		return userID > 0, nil
	}
	return s.perm.HasPermission(ctx, userID, spec.Resource, spec.Action)
}

func (s *AuthorizationService) BuildScopeFilter(ctx context.Context, userID uint, columns permission.ScopeColumns) (permission.ScopeFilter, error) {
	ds, err := s.perm.LoadDataScope(ctx, userID)
	if err != nil {
		return permission.ScopeFilter{}, err
	}

	orgID, err := s.perm.GetUserOrgID(ctx, userID)
	if err != nil {
		return permission.ScopeFilter{}, err
	}

	if ds == nil {
		return permission.ScopeFilter{}, nil
	}
	return permission.BuildDataScopeFilter(*ds, userID, orgID, columns)
}

func (s *AuthorizationService) BuildDataScopeSQL(ctx context.Context, userID uint) (string, []any, error) {
	filter, err := s.BuildScopeFilter(ctx, userID, permission.DefaultScopeColumns)
	if err != nil {
		return "", nil, err
	}
	return filter.SQL, filter.Args, nil
}

func (s *AuthorizationService) InvalidateUser(ctx context.Context, userID uint) error {
	return s.perm.InvalidateUser(ctx, userID)
}

func (s *AuthorizationService) InvalidateRole(ctx context.Context, roleID uint) error {
	return s.perm.InvalidateRoleUsers(ctx, roleID)
}

func (s *AuthorizationService) InvalidateResource(ctx context.Context, resourceID uint) error {
	return s.perm.InvalidateResourceUsers(ctx, resourceID)
}

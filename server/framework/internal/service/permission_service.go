package service

import (
	"context"
	"fmt"

	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/permission"
)

type PermissionService struct {
	permRepo   permission.UserPermissionRepository
	dsRepo     permission.DataScopeRepository
	cache      permission.PermissionCache
	platformRp permission.PlatformRoleRepository
}

func NewPermissionService(
	permRepo permission.UserPermissionRepository,
	dsRepo permission.DataScopeRepository,
	cache permission.PermissionCache,
	platformRp permission.PlatformRoleRepository,
) *PermissionService {
	return &PermissionService{
		permRepo:   permRepo,
		dsRepo:     dsRepo,
		cache:      cache,
		platformRp: platformRp,
	}
}

func (s *PermissionService) LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	if s.cache != nil {
		perms, err := s.cache.GetPermissions(ctx, userID)
		if err == nil && perms != nil {
			return perms, nil
		}
	}

	perms, err := s.permRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.SetPermissions(ctx, userID, perms)
	}

	return perms, nil
}

func (s *PermissionService) LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error) {
	if s.cache != nil {
		ds, err := s.cache.GetDataScope(ctx, userID)
		if err == nil && ds != nil {
			return ds, nil
		}
	}

	ds, err := s.dsRepo.GetDataScope(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load data scope: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.SetDataScope(ctx, userID, ds)
	}

	return ds, nil
}

func (s *PermissionService) LoadRoles(ctx context.Context, userID uint) ([]string, error) {
	return s.permRepo.GetUserRoles(ctx, userID)
}

func (s *PermissionService) InvalidateUser(ctx context.Context, userID uint) error {
	if s.cache != nil {
		_ = s.cache.InvalidatePermissions(ctx, userID)
		_ = s.cache.InvalidateDataScope(ctx, userID)
	}
	return nil
}

func (s *PermissionService) InvalidateRoleUsers(ctx context.Context, roleID uint) error {
	if s.cache == nil {
		return nil
	}

	userIDs, err := s.permRepo.GetUserIDsByRole(ctx, roleID)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		s.InvalidateUser(ctx, userID)
	}

	return nil
}

func (s *PermissionService) HasPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	perms, err := s.LoadPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	return permission.HasPermission(perms, resource, action), nil
}

func (s *PermissionService) BuildDataScopeSQL(ctx context.Context, userID uint) (string, []any, error) {
	filter, err := s.BuildDataScopeFilter(ctx, userID, permission.DefaultScopeColumns)
	if err != nil {
		return "", nil, err
	}
	return filter.SQL, filter.Args, nil
}

func (s *PermissionService) BuildDataScopeFilter(ctx context.Context, userID uint, columns permission.ScopeColumns) (permission.ScopeFilter, error) {
	ds, err := s.LoadDataScope(ctx, userID)
	if err != nil {
		return permission.ScopeFilter{}, err
	}

	orgID, err := s.dsRepo.GetUserOrgID(ctx, userID)
	if err != nil {
		return permission.ScopeFilter{}, err
	}

	if ds == nil {
		return permission.ScopeFilter{}, nil
	}
	return permission.BuildDataScopeFilter(*ds, userID, orgID, columns)
}

func (s *PermissionService) GetUserOrgID(ctx context.Context, userID uint) (int64, error) {
	return s.dsRepo.GetUserOrgID(ctx, userID)
}

func (s *PermissionService) LoadUserSecurityContext(ctx context.Context, userID uint) (perms map[string]bool, roles []string, dsPtr *permission.DataScope, orgID int64, err error) {
	var ds *permission.DataScope

	perms, err = s.LoadPermissions(ctx, userID)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	roles, err = s.LoadRoles(ctx, userID)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	ds, err = s.LoadDataScope(ctx, userID)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	if ds != nil {
		dsPtr = ds
	}

	orgID, err = s.GetUserOrgID(ctx, userID)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	// 平台超级管理员短路：忽略原本的角色/数据范围，直接授予全权限 + 全部数据
	if s.platformRp != nil {
		platformRoles, perr := s.platformRp.GetRolesByUserID(ctx, userID)
		if perr == nil {
			for _, r := range platformRoles {
				if r == jwtpkg.PlatformRoleSuperAdmin {
					perms = map[string]bool{"*:*": true}
					allDS := permission.DataScope{Type: permission.DataScopeAll}
					dsPtr = &allDS
					roles = append(roles, "super_admin")
					return perms, roles, dsPtr, orgID, nil
				}
			}
		}
		// 查询失败不阻塞常规流程
	}

	return perms, roles, dsPtr, orgID, nil
}

// LoadPlatformRoles 单独获取用户拥有的平台级角色（登录时使用）
func (s *PermissionService) LoadPlatformRoles(ctx context.Context, userID uint) []string {
	if s.platformRp == nil {
		return nil
	}
	roles, err := s.platformRp.GetRolesByUserID(ctx, userID)
	if err != nil {
		return nil
	}
	return roles
}

// IsPlatformSuperAdmin 单独判断用户是否 super_admin
func (s *PermissionService) IsPlatformSuperAdmin(ctx context.Context, userID uint) bool {
	if s.platformRp == nil {
		return false
	}
	roles, err := s.platformRp.GetRolesByUserID(ctx, userID)
	if err != nil {
		return false
	}
	for _, r := range roles {
		if r == jwtpkg.PlatformRoleSuperAdmin {
			return true
		}
	}
	return false
}

func (s *PermissionService) InvalidateResourceUsers(ctx context.Context, resourceID uint) error {
	if s.cache == nil {
		return nil
	}

	userIDs, err := s.permRepo.GetUserIDsByResource(ctx, resourceID)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		s.InvalidateUser(ctx, userID)
	}

	return nil
}

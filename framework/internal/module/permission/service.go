package permission

import (
	"gx1727.com/xin/framework/internal/module/resource"

	"gx1727.com/xin/framework/internal/module/menu"

	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	permPkg "gx1727.com/xin/framework/pkg/permission"
)

type Service struct {
	db           *pgxpool.Pool
	permRepo     permPkg.PermissionRepository
	menuRepo     menu.MenuRepository
	resourceRepo resource.ResourceRepository
}

func NewService(db *pgxpool.Pool, permRepo permPkg.PermissionRepository, menuRepo menu.MenuRepository, resourceRepo resource.ResourceRepository) *Service {
	return &Service{
		db:           db,
		permRepo:     permRepo,
		menuRepo:     menuRepo,
		resourceRepo: resourceRepo,
	}
}

func (s *Service) GetPermissions(ctx context.Context, roleID uint) (*RolePermissionsResp, error) {
	perms, err := s.permRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	resp := &RolePermissionsResp{
		Menus:     make([]MenuPerm, 0),
		Resources: make([]ResourcePerm, 0),
	}

	for _, p := range perms {
		if p.ResourceType == "menu" {
			if p.ResourceID > 0 {
				if menu, err := s.menuRepo.GetByID(ctx, p.ResourceID); err == nil {
					resp.Menus = append(resp.Menus, MenuPerm{
						ID:     menu.ID,
						Code:   menu.Code,
						Name:   menu.Name,
						Effect: p.Effect,
					})
				}
			}
		} else if p.ResourceType == "resource" {
			if p.ResourceID > 0 {
				if res, err := s.resourceRepo.GetByID(ctx, p.ResourceID); err == nil {
					resp.Resources = append(resp.Resources, ResourcePerm{
						ID:     res.ID,
						Code:   res.Code,
						Name:   res.Name,
						Action: res.Action,
						Effect: p.Effect,
					})
				}
			}
		}
	}

	return resp, nil
}

func (s *Service) AssignPermissions(ctx context.Context, tenantID, roleID uint, req AssignReq) error {
	// Delete existing permissions
	if err := s.permRepo.DeleteByRoleID(ctx, roleID); err != nil {
		return fmt.Errorf("delete existing permissions: %w", err)
	}

	// Insert new permissions
	for _, p := range req.Permissions {
		perm := permPkg.Permission{
			TenantID:     tenantID,
			RoleID:       roleID,
			ResourceType: p.ResourceType,
			ResourceID:   p.ResourceID,
			ResourceCode: p.ResourceCode,
			Effect:       p.Effect,
		}
		if err := s.permRepo.Create(ctx, tenantID, roleID, perm); err != nil {
			return fmt.Errorf("create permission: %w", err)
		}
	}

	return nil
}

func (s *Service) GetMenus(ctx context.Context, roleID uint) ([]MenuPerm, error) {
	perms, err := s.permRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	result := make([]MenuPerm, 0)
	seen := make(map[uint]bool)

	for _, p := range perms {
		if p.ResourceType != "menu" || p.ResourceID == 0 {
			continue
		}
		if seen[p.ResourceID] {
			continue
		}
		seen[p.ResourceID] = true

		if menu, err := s.menuRepo.GetByID(ctx, p.ResourceID); err == nil {
			result = append(result, MenuPerm{
				ID:     menu.ID,
				Code:   menu.Code,
				Name:   menu.Name,
				Effect: p.Effect,
			})
		}
	}

	return result, nil
}

func (s *Service) GetResources(ctx context.Context, roleID uint) ([]ResourcePerm, error) {
	perms, err := s.permRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	result := make([]ResourcePerm, 0)
	seen := make(map[uint]bool)

	for _, p := range perms {
		if p.ResourceType != "resource" || p.ResourceID == 0 {
			continue
		}
		if seen[p.ResourceID] {
			continue
		}
		seen[p.ResourceID] = true

		if res, err := s.resourceRepo.GetByID(ctx, p.ResourceID); err == nil {
			result = append(result, ResourcePerm{
				ID:     res.ID,
				Code:   res.Code,
				Name:   res.Name,
				Action: res.Action,
				Effect: p.Effect,
			})
		}
	}

	return result, nil
}

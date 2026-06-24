package permission

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/authz"
	"gx1727.com/xin/framework/pkg/xincontext"
	"gx1727.com/xin/framework/pkg/db"
)

// Service manages role-resource permission assignments via role_resources table.
// Menu permission management has moved to the role module.
type Service struct {
	db               *pgxpool.Pool
	roleResourceRepo RoleResourceRepository
	authz            authz.Authorization
}

func NewService(db *pgxpool.Pool, roleResourceRepo RoleResourceRepository, authzSvc authz.Authorization) *Service {
	return &Service{
		db:               db,
		roleResourceRepo: roleResourceRepo,
		authz:            authzSvc,
	}
}

// GetPermissions 返回角色已分配的资源权限列表
func (s *Service) GetPermissions(ctx context.Context, roleID uint) ([]ResourcePerm, error) {
	return s.GetResources(ctx, roleID)
}

// AssignPermissions 全量覆盖角色的资源权限
func (s *Service) AssignPermissions(ctx context.Context, tenantID, roleID uint, req AssignResourceReq) error {
	err := db.RunInTenantTx(ctx, s.db, tenantID, func(ctx context.Context) error {
		return s.roleResourceRepo.SetForRole(ctx, roleID, req.ResourceIDs)
	})
	if err != nil {
		return err
	}

	// 使关联该角色的用户缓存失效
	if s.authz != nil {
		_ = s.authz.InvalidateRole(context.Background(), roleID)
	}

	return nil
}

// GetResources 查询角色的资源权限列表
func (s *Service) GetResources(ctx context.Context, roleID uint) ([]ResourcePerm, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)

	result := make([]ResourcePerm, 0)

	err := db.RunInTenantTx(ctx, s.db, tenantID, func(ctx context.Context) error {
		resourceIDs, err := s.roleResourceRepo.GetByRoleID(ctx, roleID)
		if err != nil {
			return err
		}

		for _, resID := range resourceIDs {
			// 直接查 resources 表获取详情
			var (
				id, menuID         uint
				code, name, action string
			)
			q, qErr := db.GetQuerier(ctx, s.db)
			if qErr != nil {
				continue
			}
			err := q.QueryRow(ctx, `
				SELECT id, menu_id, code, name, action FROM tenant_permissions
				WHERE is_deleted = FALSE AND id = $1`, resID).Scan(&id, &menuID, &code, &name, &action)
			if err != nil {
				continue // 跳过已删除的资源
			}
			result = append(result, ResourcePerm{
				ID:     id,
				Code:   code,
				Name:   name,
				Action: action,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

package role

import (
	"context"

	"gx1727.com/xin/framework/internal/service"
	"gx1727.com/xin/framework/pkg/cache"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
)

type Service struct {
	roleRepo RoleRepository
	dsRepo   permission.DataScopeRepository
	menuRepo RoleMenuRepository
}

func NewService(roleRepo RoleRepository, dsRepo permission.DataScopeRepository, menuRepo RoleMenuRepository) *Service {
	return &Service{roleRepo: roleRepo, dsRepo: dsRepo, menuRepo: menuRepo}
}

func (s *Service) List(ctx context.Context, tenantID uint, req ListReq) ([]RoleResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}
	var roles []Role
	var total int64
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		roles, total, err = s.roleRepo.List(ctx, tenantID, req.Keyword, req.Page, req.Size)
		return err
	})
	if err != nil {
		return nil, 0, err
	}
	result := make([]RoleResp, len(roles))
	for i, r := range roles {
		result[i] = toResp(r)
	}
	return result, total, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*RoleResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var role *Role
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		role, err = s.roleRepo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	resp := toResp(*role)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req CreateReq) (*RoleResp, error) {
	if req.Status == 0 {
		req.Status = 1
	}
	var role *Role
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		role, err = s.roleRepo.Create(ctx, tenantID, CreateRoleRepoReq{
			Code:        req.Code,
			Name:        req.Name,
			Description: req.Description,
			DataScope:   req.DataScope,
			IsDefault:   req.IsDefault,
			Sort:        req.Sort,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	resp := toResp(*role)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*RoleResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var role *Role
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		role, err = s.roleRepo.Update(ctx, id, UpdateRoleRepoReq{
			Name:        req.Name,
			Description: req.Description,
			DataScope:   req.DataScope,
			IsDefault:   req.IsDefault,
			Sort:        req.Sort,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	if cache.Get() != nil {
		permService := service.NewPermissionService(
			permission.NewPermissionRepository(db.Get()),
			permission.NewDataScopeRepository(db.Get()),
			permission.NewRedisPermissionCache(),
		)
		_ = permService.InvalidateRoleUsers(context.Background(), id)
	}

	resp := toResp(*role)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	tenantID, _ := xincontext.TenantIDFrom(ctx)

	// Get users before deleting the role so we can invalidate them
	var userIDs []uint
	if cache.Get() != nil {
		permRepo := permission.NewPermissionRepository(db.Get())
		userIDs, _ = permRepo.GetUserIDsByRole(ctx, id)
	}

	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		role, err := s.roleRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if role.Code == "admin" {
			return ErrCannotDeleteAdmin
		}
		return s.roleRepo.Delete(ctx, id)
	})
	if err != nil {
		return err
	}

	if cache.Get() != nil && len(userIDs) > 0 {
		permService := service.NewPermissionService(
			permission.NewPermissionRepository(db.Get()),
			permission.NewDataScopeRepository(db.Get()),
			permission.NewRedisPermissionCache(),
		)
		for _, uid := range userIDs {
			_ = permService.InvalidateUser(context.Background(), uid)
		}
	}

	return nil
}

func (s *Service) GetDataScopes(ctx context.Context, roleID uint) (*DataScopeResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var orgIDs []uint
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		orgIDs, err = s.dsRepo.GetByRoleID(ctx, roleID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &DataScopeResp{OrgIDs: orgIDs}, nil
}

func (s *Service) UpdateDataScopes(ctx context.Context, roleID uint, req UpdateDataScopesReq) error {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		return s.dsRepo.SetForRole(ctx, roleID, req.OrgIDs)
	})
	if err != nil {
		return err
	}

	if cache.Get() != nil {
		permService := service.NewPermissionService(
			permission.NewPermissionRepository(db.Get()),
			permission.NewDataScopeRepository(db.Get()),
			permission.NewRedisPermissionCache(),
		)
		_ = permService.InvalidateRoleUsers(context.Background(), roleID)
	}

	return nil
}

// GetMenus 获取角色的菜单权限
func (s *Service) GetMenus(ctx context.Context, roleID uint) (*RoleMenuResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var menuIDs []uint
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		menuIDs, err = s.menuRepo.GetByRoleID(ctx, roleID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &RoleMenuResp{MenuIDs: menuIDs}, nil
}

// AssignMenus 分配角色的菜单权限（全量覆盖）
func (s *Service) AssignMenus(ctx context.Context, roleID uint, req AssignMenusReq) error {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		return s.menuRepo.SetForRole(ctx, roleID, req.MenuIDs)
	})
}

func toResp(r Role) RoleResp {
	return RoleResp{
		ID:          r.ID,
		TenantID:    r.TenantID,
		OrgID:       r.OrgID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		DataScope:   r.DataScope,
		Extend:      r.Extend,
		IsDefault:   r.IsDefault,
		Sort:        r.Sort,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

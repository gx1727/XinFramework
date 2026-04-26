package role

import (
	"context"

	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/permission"
)

type Service struct {
	roleRepo model.RoleRepository
	dsRepo   permission.DataScopeRepository
}

func NewService(roleRepo model.RoleRepository, dsRepo permission.DataScopeRepository) *Service {
	return &Service{roleRepo: roleRepo, dsRepo: dsRepo}
}

func (s *Service) List(ctx context.Context, tenantID uint, req ListReq) ([]RoleResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}
	roles, total, err := s.roleRepo.List(ctx, tenantID, req.Keyword, req.Page, req.Size)
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
	role, err := s.roleRepo.GetByID(ctx, id)
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
	role, err := s.roleRepo.Create(ctx, tenantID, model.CreateRoleRepoReq{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		DataScope:   req.DataScope,
		IsDefault:   req.IsDefault,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*role)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*RoleResp, error) {
	role, err := s.roleRepo.Update(ctx, id, model.UpdateRoleRepoReq{
		Name:        req.Name,
		Description: req.Description,
		DataScope:   req.DataScope,
		IsDefault:   req.IsDefault,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*role)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.Code == "admin" {
		return ErrCannotDeleteAdmin
	}
	return s.roleRepo.Delete(ctx, id)
}

func (s *Service) GetDataScopes(ctx context.Context, roleID uint) (*DataScopeResp, error) {
	orgIDs, err := s.dsRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return &DataScopeResp{OrgIDs: orgIDs}, nil
}

func (s *Service) UpdateDataScopes(ctx context.Context, roleID uint, req UpdateDataScopesReq) error {
	return s.dsRepo.SetForRole(ctx, roleID, req.OrgIDs)
}

func toResp(r model.Role) RoleResp {
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

package resource

import (
	"context"

	"gx1727.com/xin/framework/internal/service"
)

type Service struct {
	resourceRepo ResourceRepository
}

func NewService(resourceRepo ResourceRepository) *Service {
	return &Service{resourceRepo: resourceRepo}
}

func (s *Service) List(ctx context.Context, tenantID uint, req ListReq) ([]ResourceResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	var resources []Resource
	var total int64
	var err error
	if req.MenuID > 0 {
		resources, err = s.resourceRepo.GetByMenu(ctx, req.MenuID)
	} else {
		resources, err = s.resourceRepo.GetByTenant(ctx, tenantID)
	}
	if err != nil {
		return nil, 0, err
	}
	total = int64(len(resources))

	result := make([]ResourceResp, 0, len(resources))
	for _, r := range resources {
		if req.Action != "" && r.Action != req.Action {
			continue
		}
		result = append(result, toResp(r))
	}

	return result, total, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*ResourceResp, error) {
	r, err := s.resourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req CreateReq) (*ResourceResp, error) {
	if req.Status == 0 {
		req.Status = 1
	}
	r, err := s.resourceRepo.Create(ctx, tenantID, CreateResourceRepoReq{
		MenuID:      req.MenuID,
		Code:        req.Code,
		Name:        req.Name,
		Action:      req.Action,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*ResourceResp, error) {
	r, err := s.resourceRepo.Update(ctx, id, UpdateResourceRepoReq{
		Name:        req.Name,
		Action:      req.Action,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}

	if ps := service.GlobalPermissionService(); ps != nil {
		_ = ps.InvalidateResourceUsers(context.Background(), id)
	}

	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if ps := service.GlobalPermissionService(); ps != nil {
		_ = ps.InvalidateResourceUsers(context.Background(), id)
	}

	return s.resourceRepo.Delete(ctx, id)
}

func (s *Service) GetByMenu(ctx context.Context, menuID uint) ([]ResourceResp, error) {
	resources, err := s.resourceRepo.GetByMenu(ctx, menuID)
	if err != nil {
		return nil, err
	}
	result := make([]ResourceResp, len(resources))
	for i, r := range resources {
		result[i] = toResp(r)
	}
	return result, nil
}

func (s *Service) GetUserResourcesByMenu(ctx context.Context, tenantID, userID, menuID uint) ([]ResourceResp, error) {
	resources, err := s.resourceRepo.GetUserResourcesByMenu(ctx, tenantID, userID, menuID)
	if err != nil {
		return nil, err
	}
	result := make([]ResourceResp, len(resources))
	for i, r := range resources {
		result[i] = toResp(r)
	}
	return result, nil
}

func toResp(r Resource) ResourceResp {
	return ResourceResp{
		ID:          r.ID,
		TenantID:    r.TenantID,
		MenuID:      r.MenuID,
		Code:        r.Code,
		Name:        r.Name,
		Action:      r.Action,
		Description: r.Description,
		Sort:        r.Sort,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

package tenant

import (
	"context"

	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	tenantRepo model.TenantRepository
}

func NewService(repo model.TenantRepository) *Service {
	return &Service{tenantRepo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uint) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	t, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	resp := toResp(t)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, req CreateTenantReq) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	t, err := s.tenantRepo.Create(ctx, req.Code, req.Name, req.Contact, req.Phone, req.Email)
	if err != nil {
		return nil, mapRepoError(err)
	}
	resp := toResp(t)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateTenantReq) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	t, err := s.tenantRepo.Update(ctx, id, req.Name, req.Contact, req.Phone, req.Email,
		req.Province, req.City, req.Area, req.Address)
	if err != nil {
		return nil, mapRepoError(err)
	}
	resp := toResp(t)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.tenantRepo == nil {
		return ErrBackendUnavailable
	}
	return mapRepoError(s.tenantRepo.Delete(ctx, id))
}

func (s *Service) List(ctx context.Context, req ListTenantReq) ([]TenantResp, int64, error) {
	if s.tenantRepo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	list, total, err := s.tenantRepo.List(ctx, req.Keyword, req.Status, req.Page, req.Size)
	if err != nil {
		return nil, 0, mapRepoError(err)
	}
	resps := make([]TenantResp, len(list))
	for i := range list {
		resps[i] = toResp(&list[i])
	}
	return resps, total, nil
}

package tenant

import (
	"context"

	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	tenantRepo TenantRepository
}

func NewService(repo TenantRepository) *Service {
	return &Service{tenantRepo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uint) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var t *Tenant
	err := db.RunInTx(ctx, db.Get(), func(ctx context.Context) error {
		var err error
		t, err = s.tenantRepo.GetByID(ctx, id)
		return err
	})
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
	var t *Tenant
	err := db.RunInTx(ctx, db.Get(), func(ctx context.Context) error {
		var err error
		t, err = s.tenantRepo.Create(ctx, req.Code, req.Name, req.Contact, req.Phone, req.Email)
		return err
	})
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
	var t *Tenant
	err := db.RunInTx(ctx, db.Get(), func(ctx context.Context) error {
		var err error
		t, err = s.tenantRepo.Update(ctx, id, req.Name, req.Contact, req.Phone, req.Email,
			req.Province, req.City, req.Area, req.Address)
		return err
	})
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
	err := db.RunInTx(ctx, db.Get(), func(ctx context.Context) error {
		return s.tenantRepo.Delete(ctx, id)
	})
	return mapRepoError(err)
}

func (s *Service) List(ctx context.Context, req ListTenantReq) ([]TenantResp, int64, error) {
	if s.tenantRepo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	var list []Tenant
	var total int64
	err := db.RunInTx(ctx, db.Get(), func(ctx context.Context) error {
		var err error
		list, total, err = s.tenantRepo.List(ctx, req.Keyword, req.Status, req.Page, req.Size)
		return err
	})
	if err != nil {
		return nil, 0, mapRepoError(err)
	}
	resps := make([]TenantResp, len(list))
	for i := range list {
		resps[i] = toResp(&list[i])
	}
	return resps, total, nil
}

package tenant

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetByID(ctx context.Context, id uint) (*TenantResp, error) {
	if s.db == nil {
		return nil, ErrBackendUnavailable
	}
	t, err := GetByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	resp := t.ToResp()
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, req CreateTenantReq) (*TenantResp, error) {
	if s.db == nil {
		return nil, ErrBackendUnavailable
	}
	t, err := Create(ctx, s.db, req)
	if err != nil {
		return nil, err
	}
	resp := t.ToResp()
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateTenantReq) (*TenantResp, error) {
	if s.db == nil {
		return nil, ErrBackendUnavailable
	}
	t, err := Update(ctx, s.db, id, req)
	if err != nil {
		return nil, err
	}
	resp := t.ToResp()
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.db == nil {
		return ErrBackendUnavailable
	}
	return Delete(ctx, s.db, id)
}

func (s *Service) List(ctx context.Context, req ListTenantReq) ([]TenantResp, int64, error) {
	if s.db == nil {
		return nil, 0, ErrBackendUnavailable
	}
	list, total, err := List(ctx, s.db, req)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]TenantResp, len(list))
	for i := range list {
		resps[i] = list[i].ToResp()
	}
	return resps, total, nil
}

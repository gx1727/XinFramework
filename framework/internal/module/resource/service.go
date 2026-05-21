package resource

import (
	"gx1727.com/xin/framework/internal/module/menu"

	"context"

	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	resourceRepo ResourceRepository
	menuRepo     menu.MenuRepository
}

func NewService(resourceRepo ResourceRepository, menuRepo menu.MenuRepository) *Service {
	return &Service{resourceRepo: resourceRepo, menuRepo: menuRepo}
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

	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		if req.MenuID > 0 {
			resources, err = s.resourceRepo.GetByMenu(ctx, req.MenuID)
			total = int64(len(resources))
		} else {
			resources, err = s.resourceRepo.GetByTenant(ctx, tenantID)
			total = int64(len(resources))
		}
		return err
	})

	if err != nil {
		return nil, 0, err
	}

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
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var r *Resource
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		r, err = s.resourceRepo.GetByID(ctx, id)
		return err
	})
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
	var r *Resource
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		r, err = s.resourceRepo.Create(ctx, tenantID, CreateResourceRepoReq{
			MenuID:      req.MenuID,
			Code:        req.Code,
			Name:        req.Name,
			Action:      req.Action,
			Description: req.Description,
			Sort:        req.Sort,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*ResourceResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var r *Resource
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		r, err = s.resourceRepo.Update(ctx, id, UpdateResourceRepoReq{
			Name:        req.Name,
			Action:      req.Action,
			Description: req.Description,
			Sort:        req.Sort,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		return s.resourceRepo.Delete(ctx, id)
	})
}

func (s *Service) GetByMenu(ctx context.Context, menuID uint) ([]ResourceResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var resources []Resource
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		resources, err = s.resourceRepo.GetByMenu(ctx, menuID)
		return err
	})
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

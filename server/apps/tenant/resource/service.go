package resource

import (
	"context"
	"strings"

	"gx1727.com/xin/framework/pkg/authz"
)

type Service struct {
	resourceRepo ResourceRepository
	authz        authz.Authorization
}

func NewService(resourceRepo ResourceRepository, authzSvc authz.Authorization) *Service {
	return &Service{
		resourceRepo: resourceRepo,
		authz:        authzSvc,
	}
}

// permissionCodeValid 校验权限码格式：resource:action（仅含一个冒号，前后非空）。
// 与 apps/platform/sys_permission/service.go permissionCodeValid 规则一致；
// 0024 统一约定后，tenant_permissions.code 必须是完整串，不再两段式。
func permissionCodeValid(code string) bool {
	idx := strings.Index(code, ":")
	if idx <= 0 || idx == len(code)-1 {
		return false
	}
	if strings.Count(code, ":") != 1 {
		return false
	}
	return true
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
		return nil, mapRepoError(err)
	}
	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req CreateReq) (*ResourceResp, error) {
	if !permissionCodeValid(req.Code) {
		return nil, ErrResourceInvalidCode
	}
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
		return nil, mapRepoError(err)
	}
	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*ResourceResp, error) {
	if req.Code != nil && !permissionCodeValid(*req.Code) {
		return nil, ErrResourceInvalidCode
	}
	r, err := s.resourceRepo.Update(ctx, id, UpdateResourceRepoReq{
		Code:        req.Code,
		Name:        req.Name,
		Action:      req.Action,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	if s.authz != nil {
		_ = s.authz.InvalidateResource(context.Background(), id)
	}

	resp := toResp(*r)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.authz != nil {
		_ = s.authz.InvalidateResource(context.Background(), id)
	}

	return mapRepoError(s.resourceRepo.Delete(ctx, id))
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

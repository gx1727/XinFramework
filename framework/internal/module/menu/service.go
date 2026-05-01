package menu

import (
	"context"
	"errors"
)

type Service struct {
	menuRepo MenuRepository
}

func NewService(repo MenuRepository) *Service {
	return &Service{menuRepo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uint) (*MenuResp, error) {
	if s.menuRepo == nil {
		return nil, ErrBackendUnavailable
	}
	m, err := s.menuRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMenuNotFoundDB) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}
	return toResp(m), nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req CreateMenuReq) (*MenuResp, error) {
	if s.menuRepo == nil {
		return nil, ErrBackendUnavailable
	}

	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	repoReq := CreateMenuRepoReq{
		Code:      req.Code,
		Name:      req.Name,
		Subtitle:  req.Subtitle,
		URL:       req.URL,
		Path:      req.Path,
		Icon:      req.Icon,
		Sort:      req.Sort,
		ParentID:  req.ParentID,
		Ancestors: req.Ancestors,
		Visible:   visible,
		Enabled:   enabled,
	}

	m, err := s.menuRepo.Create(ctx, tenantID, repoReq)
	if err != nil {
		if err.Error() == "menu code already exists" {
			return nil, ErrMenuCodeExists
		}
		return nil, err
	}
	return toResp(m), nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateMenuReq) (*MenuResp, error) {
	if s.menuRepo == nil {
		return nil, ErrBackendUnavailable
	}

	repoReq := UpdateMenuRepoReq{
		Code:     req.Code,
		Name:     req.Name,
		Subtitle: req.Subtitle,
		URL:      req.URL,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Visible:  true,
		Enabled:  true,
	}
	if req.Visible != nil {
		repoReq.Visible = *req.Visible
	}
	if req.Enabled != nil {
		repoReq.Enabled = *req.Enabled
	}

	m, err := s.menuRepo.Update(ctx, id, repoReq)
	if err != nil {
		if errors.Is(err, ErrMenuNotFoundDB) {
			return nil, ErrMenuNotFound
		}
		if err.Error() == "menu code already exists" {
			return nil, ErrMenuCodeExists
		}
		return nil, err
	}
	return toResp(m), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.menuRepo == nil {
		return ErrBackendUnavailable
	}
	err := s.menuRepo.Delete(ctx, id)
	if errors.Is(err, ErrMenuNotFoundDB) {
		return ErrMenuNotFound
	}
	return err
}

func (s *Service) List(ctx context.Context, tenantID uint, req ListMenuReq) ([]MenuResp, int64, error) {
	if s.menuRepo == nil {
		return nil, 0, ErrBackendUnavailable
	}

	menus, err := s.menuRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Filter root menus if requested
	if req.Root {
		var filtered []Menu
		for _, m := range menus {
			if m.ParentID == 0 {
				filtered = append(filtered, m)
			}
		}
		menus = filtered
	}

	resps := make([]MenuResp, len(menus))
	for i, m := range menus {
		resps[i] = *toResp(&m)
	}

	return resps, int64(len(resps)), nil
}

func (s *Service) Tree(ctx context.Context, tenantID uint) ([]*MenuResp, error) {
	if s.menuRepo == nil {
		return nil, ErrBackendUnavailable
	}

	menus, err := s.menuRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return buildTree(menus), nil
}

// buildTree builds a tree structure from flat menu list
func buildTree(menus []Menu) []*MenuResp {
	// Create map for quick lookup
	nodeMap := make(map[uint]*MenuResp)
	for _, m := range menus {
		nodeMap[m.ID] = toResp(&m)
	}

	// Build tree
	var roots []*MenuResp
	for _, m := range menus {
		node := nodeMap[m.ID]
		if m.ParentID == 0 {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[m.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots
}

func toResp(m *Menu) *MenuResp {
	if m == nil {
		return nil
	}
	return &MenuResp{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Code:      m.Code,
		Name:      m.Name,
		Subtitle:  m.Subtitle,
		URL:       m.URL,
		Path:      m.Path,
		Icon:      m.Icon,
		Sort:      m.Sort,
		ParentID:  m.ParentID,
		Ancestors: m.Ancestors,
		Visible:   m.Visible,
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrMenuNotFoundDB) {
		return ErrMenuNotFound
	}
	if err.Error() == "menu code already exists" {
		return ErrMenuCodeExists
	}
	return err
}

func MapRespError(err error) int {
	if errors.Is(err, ErrMenuNotFound) {
		return 404
	}
	if errors.Is(err, ErrMenuCodeExists) {
		return 400
	}
	if errors.Is(err, ErrBackendUnavailable) {
		return 500
	}
	return 500
}

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
		Code:      req.Code,
		Name:      req.Name,
		Subtitle:  req.Subtitle,
		URL:       req.URL,
		Path:      req.Path,
		Icon:      req.Icon,
		Sort:      req.Sort,
		ParentID:  req.ParentID,
		Ancestors: req.Ancestors,
		Visible:   true,
		Enabled:   true,
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

func (s *Service) Delete(ctx context.Context, tenantID uint, id uint) error {
	if s.menuRepo == nil {
		return ErrBackendUnavailable
	}
	err := s.menuRepo.Delete(ctx, tenantID, id)
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

	if req.Root {
		var filtered []Menu
		for _, m := range menus {
			if m.ParentID == nil || *m.ParentID == 0 {
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
// Uses DFS-based cycle detection to prevent infinite loops in case of circular parent-child references
func buildTree(menus []Menu) []*MenuResp {
	nodeMap := make(map[uint]*MenuResp)
	for _, m := range menus {
		nodeMap[m.ID] = toResp(&m)
	}

	// Phase 1: attach each node to its parent (if parent exists in this dataset)
	var roots []*MenuResp
	for _, m := range menus {
		node := nodeMap[m.ID]
		if m.ParentID == nil || *m.ParentID == 0 {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*m.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				roots = append(roots, node)
			}
		}
	}

	// Phase 2: break all cycles using DFS with a global visited set
	// Any node that appears as its own ancestor (via parent chain) is moved to roots
	visited := make(map[uint]int) // 0=unvisited, 1=visiting, 2=done
	var breakCycle func(nodeID uint) bool
	breakCycle = func(nodeID uint) bool {
		if nodeID == 0 {
			return false
		}
		state, ok := visited[nodeID]
		if !ok {
			visited[nodeID] = 1
		} else if state == 1 {
			// Found a back-edge: cycle detected — move this node (and its subtree) to roots
			if node, exists := nodeMap[nodeID]; exists {
				removeFromParent(node, roots)
				if !inRoots(node, roots) {
					roots = append(roots, node)
				}
			}
			return true
		} else {
			return false
		}

		if node, ok := nodeMap[nodeID]; ok && node.ParentID != nil && *node.ParentID != 0 {
			breakCycle(*node.ParentID)
		}
		visited[nodeID] = 2
		return false
	}

	for _, m := range menus {
		breakCycle(m.ID)
	}

	return roots
}

func removeFromParent(node *MenuResp, roots []*MenuResp) {
	for _, root := range roots {
		for j := len(root.Children) - 1; j >= 0; j-- {
			if root.Children[j].ID == node.ID {
				root.Children = append(root.Children[:j], root.Children[j+1:]...)
				return
			}
			removeFromParentDeep(root.Children[j], node, roots)
		}
	}
}

func removeFromParentDeep(parent *MenuResp, target *MenuResp, roots []*MenuResp) {
	for j := len(parent.Children) - 1; j >= 0; j-- {
		if parent.Children[j].ID == target.ID {
			parent.Children = append(parent.Children[:j], parent.Children[j+1:]...)
			return
		}
		removeFromParentDeep(parent.Children[j], target, roots)
	}
}

func inRoots(node *MenuResp, roots []*MenuResp) bool {
	for _, r := range roots {
		if r.ID == node.ID {
			return true
		}
	}
	return false
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
	if errors.Is(err, ErrMenuCodeExistsDB) {
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

package organization

import (
	"context"
	"fmt"
	"strings"

	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	orgRepo model.OrganizationRepository
}

func NewService(orgRepo model.OrganizationRepository) *Service {
	return &Service{orgRepo: orgRepo}
}

func (s *Service) List(ctx context.Context, tenantID uint, req ListReq) ([]OrgResp, int64, error) {
	var orgs []model.Organization
	var err error

	if req.ParentID > 0 {
		orgs, err = s.orgRepo.GetChildren(ctx, req.ParentID)
	} else {
		orgs, err = s.orgRepo.GetByTenant(ctx, tenantID)
	}
	if err != nil {
		return nil, 0, err
	}

	// Filter by keyword in-memory for now
	result := make([]OrgResp, 0, len(orgs))
	for _, org := range orgs {
		if req.Keyword != "" {
			if !strings.Contains(org.Name, req.Keyword) && !strings.Contains(org.Code, req.Keyword) {
				continue
			}
		}
		result = append(result, toResp(org))
	}

	return result, int64(len(result)), nil
}

func (s *Service) Get(ctx context.Context, id uint) (*OrgResp, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toResp(*org)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, tenantID uint, req CreateReq) (*OrgResp, error) {
	if req.Status == 0 {
		req.Status = 1
	}

	// Build ancestors path
	ancestors := fmt.Sprintf("%d", req.ParentID)
	if req.ParentID > 0 {
		parent, err := s.orgRepo.GetByID(ctx, req.ParentID)
		if err == nil && parent.Ancestors != "" {
			ancestors = parent.Ancestors + "." + ancestors
		}
	}

	org, err := s.orgRepo.Create(ctx, tenantID, model.CreateOrgRepoReq{
		Code:        req.Code,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		AdminCode:   req.AdminCode,
		ParentID:    req.ParentID,
		Ancestors:   ancestors,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*org)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*OrgResp, error) {
	org, err := s.orgRepo.Update(ctx, id, model.UpdateOrgRepoReq{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		AdminCode:   req.AdminCode,
		Sort:        req.Sort,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*org)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if org.ParentID == 0 {
		return ErrCannotDeleteRoot
	}
	return s.orgRepo.Delete(ctx, id)
}

func (s *Service) GetTree(ctx context.Context, tenantID uint) ([]OrgResp, error) {
	orgs, err := s.orgRepo.GetTree(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return buildTree(orgs), nil
}

func buildTree(orgs []model.Organization) []OrgResp {
	orgMap := make(map[uint]*OrgResp)
	var roots []OrgResp

	// First pass: create all nodes
	for _, org := range orgs {
		node := toResp(org)
		node.Children = make([]OrgResp, 0)
		orgMap[org.ID] = &node
	}

	// Second pass: build tree
	for _, org := range orgs {
		node := orgMap[org.ID]
		if org.ParentID == 0 {
			roots = append(roots, *node)
		} else {
			if parent, ok := orgMap[org.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	// Convert pointers back to values
	result := make([]OrgResp, len(roots))
	for i, root := range roots {
		result[i] = root
	}

	return result
}

func toResp(m model.Organization) OrgResp {
	return OrgResp{
		ID:          m.ID,
		TenantID:    m.TenantID,
		Code:        m.Code,
		Name:        m.Name,
		Type:        m.Type,
		Description: m.Description,
		AdminCode:   m.AdminCode,
		ParentID:    m.ParentID,
		Ancestors:   m.Ancestors,
		Sort:        m.Sort,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

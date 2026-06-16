package organization

import (
	"context"
	"fmt"
	"strings"

	"gx1727.com/xin/framework/pkg/audit"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	orgRepo OrganizationRepository
}

func NewService(orgRepo OrganizationRepository) *Service {
	return &Service{orgRepo: orgRepo}
}

func (s *Service) List(ctx context.Context, tenantID uint, req ListReq) ([]OrgResp, int64, error) {
	var orgs []Organization

	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		if req.ParentID > 0 {
			orgs, err = s.orgRepo.GetChildrenScoped(ctx, req.ParentID)
		} else {
			orgs, err = s.orgRepo.GetByTenantScoped(ctx, tenantID)
		}
		return err
	})

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
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var org *Organization
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		org, err = s.orgRepo.GetByIDScoped(ctx, id)
		return err
	})
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

	var org *Organization
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		// Build ancestors path
		ancestors := fmt.Sprintf("%d", req.ParentID)
		if req.ParentID > 0 {
			parent, err := s.orgRepo.GetByIDScoped(ctx, req.ParentID)
			if err == nil && parent.Ancestors != "" {
				ancestors = parent.Ancestors + "." + ancestors
			}
		}

		var err error
		org, err = s.orgRepo.Create(ctx, tenantID, CreateOrgRepoReq{
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
		return err
	})

	if err != nil {
		return nil, err
	}
	resp := toResp(*org)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*OrgResp, error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	var org *Organization
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		if _, err := s.orgRepo.GetByIDScoped(ctx, id); err != nil {
			return err
		}
		var err error
		org, err = s.orgRepo.Update(ctx, id, UpdateOrgRepoReq{
			Name:        req.Name,
			Type:        req.Type,
			Description: req.Description,
			AdminCode:   req.AdminCode,
			Sort:        req.Sort,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	resp := toResp(*org)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		org, err := s.orgRepo.GetByIDScoped(ctx, id)
		if err != nil {
			return err
		}
		if org.ParentID == 0 {
			return ErrCannotDeleteRoot
		}

		// 0) 子组织还在，禁止删除
		children, err := s.orgRepo.CountChildren(ctx, id)
		if err != nil {
			return fmt.Errorf("count children: %w", err)
		}
		if children > 0 {
			return fmt.Errorf("%w (子组织数=%d)", ErrOrgHasUsers, children)
		}

		// 1) 子树下还有用户，禁止删除。用 org 仓库自带的 CountUsersInOrgTree，跨表查询在仓库内封闭。
		n, err := s.orgRepo.CountUsersInOrgTree(ctx, id)
		if err != nil {
			return fmt.Errorf("count org users: %w", err)
		}
		if n > 0 {
			return fmt.Errorf("%w (用户数=%d)", ErrOrgHasUsers, n)
		}

		if err := s.orgRepo.Delete(ctx, id); err != nil {
			return err
		}

		// 2) 审计：记录组织软删除事件
		audit.Log(ctx, audit.Entry{
			TenantID:  org.TenantID,
			Action:    "org:delete",
			TableName: "organizations",
			RecordID:  org.ID,
			OldData: map[string]any{
				"id":        org.ID,
				"code":      org.Code,
				"name":      org.Name,
				"type":      org.Type,
				"parent_id": org.ParentID,
				"ancestors": org.Ancestors,
				"status":    org.Status,
			},
		})
		return nil
	})
}

func (s *Service) GetTree(ctx context.Context, tenantID uint) ([]OrgResp, error) {
	var orgs []Organization
	err := db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
		var err error
		orgs, err = s.orgRepo.GetTreeScoped(ctx, tenantID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return buildTree(orgs), nil
}

func buildTree(orgs []Organization) []OrgResp {
	type tnode struct {
		resp     OrgResp
		children []*tnode
	}

	nodes := make(map[uint]*tnode, len(orgs))
	for i := range orgs {
		org := orgs[i]
		node := toResp(org)
		node.Children = []OrgResp{}
		nodes[org.ID] = &tnode{resp: node}
	}

	var roots []*tnode
	for i := range orgs {
		org := orgs[i]
		n, ok := nodes[org.ID]
		if !ok {
			continue
		}
		if org.ParentID == 0 {
			roots = append(roots, n)
			continue
		}
		if parent, ok := nodes[org.ParentID]; ok {
			parent.children = append(parent.children, n)
		}
	}

	var toList func(ns []*tnode) []OrgResp
	toList = func(ns []*tnode) []OrgResp {
		out := make([]OrgResp, 0, len(ns))
		for _, n := range ns {
			r := n.resp
			r.Children = toList(n.children)
			out = append(out, r)
		}
		return out
	}

	return toList(roots)
}

func toResp(m Organization) OrgResp {
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

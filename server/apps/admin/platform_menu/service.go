package platformmenu

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

// Service 是 platform_menu 的业务逻辑层。
//
// 关键不变量（在所有方法里强制）：
//
//  1. 所有 DB 操作必须走 db.RunInPlatformTx（绕过 RLS）
//  2. 所有 SQL 强制 tenant_id = 0（由 repository 内部常量保证）
//  3. Service **不接收** tenantID 参数 —— 平台菜单不存在"哪个租户"概念
//
// 这三个不变量是 platform_menu vs rbac/menu 的本质区别。
// 任何未来修改如果违反这些约束，应当在 code review 直接拒掉。
type Service struct {
	pool *pgxpool.Pool
	repo MenuRepository
}

func NewService(pool *pgxpool.Pool, repo MenuRepository) *Service {
	return &Service{pool: pool, repo: repo}
}

// List 列出全部平台菜单（平台菜单数量小，不分页）。
func (s *Service) List(ctx context.Context) ([]*MenuResp, int64, error) {
	if s.repo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	var out []*MenuResp
	var total int64
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		menus, err := s.repo.GetAll(txCtx)
		if err != nil {
			return err
		}
		out = make([]*MenuResp, len(menus))
		for i := range menus {
			out[i] = toResp(&menus[i])
		}
		total = int64(len(out))
		return nil
	})
	return out, total, err
}

// Tree 返回平台菜单树。
func (s *Service) Tree(ctx context.Context) ([]*MenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var tree []*MenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		menus, err := s.repo.GetAll(txCtx)
		if err != nil {
			return err
		}
		tree = buildTree(menus)
		return nil
	})
	return tree, err
}

// GetByID 按 ID 取一个平台菜单。
func (s *Service) GetByID(ctx context.Context, id uint) (*MenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var out *MenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		m, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		out = toResp(m)
		return nil
	})
	if err != nil {
		if errors.Is(err, errMenuNotFoundDB) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}
	return out, nil
}

// Create 新建平台菜单。
//
// tenant_id 在 repository 层硬编码为 0，调用方无法影响。
func (s *Service) Create(ctx context.Context, req CreateMenuReq) (*MenuResp, error) {
	if s.repo == nil {
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

	repoReq := CreateRepoReq{
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

	var out *MenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		m, err := s.repo.Create(txCtx, repoReq)
		if err != nil {
			return err
		}
		out = toResp(m)
		return nil
	})
	if err != nil {
		if err.Error() == "menu code already exists" {
			return nil, ErrMenuCodeExists
		}
		return nil, err
	}
	return out, nil
}

// Update 修改平台菜单。
func (s *Service) Update(ctx context.Context, id uint, req UpdateMenuReq) (*MenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}

	repoReq := UpdateRepoReq{
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

	var out *MenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		m, err := s.repo.Update(txCtx, id, repoReq)
		if err != nil {
			return err
		}
		out = toResp(m)
		return nil
	})
	if err != nil {
		if errors.Is(err, errMenuNotFoundDB) {
			return nil, ErrMenuNotFound
		}
		if err.Error() == "menu code already exists" {
			return nil, ErrMenuCodeExists
		}
		return nil, err
	}
	return out, nil
}

// Delete 软删除平台菜单。
func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id)
	})
	if err != nil {
		if errors.Is(err, errMenuNotFoundDB) {
			return ErrMenuNotFound
		}
		return err
	}
	return nil
}

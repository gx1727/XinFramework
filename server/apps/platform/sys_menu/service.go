package sysmenu

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	pool *pgxpool.Pool
	repo Repository
}

func NewService(pool *pgxpool.Pool, repo Repository) *Service {
	return &Service{pool: pool, repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*SysMenuResp, int64, error) {
	if s.repo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	var out []*SysMenuResp
	var total int64
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		menus, err := s.repo.GetAll(txCtx)
		if err != nil {
			return err
		}
		total = int64(len(menus))
		out = make([]*SysMenuResp, len(menus))
		for i := range menus {
			out[i] = toResp(&menus[i])
		}
		return nil
	})
	return out, total, err
}

func (s *Service) Tree(ctx context.Context) ([]*SysMenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var tree []*SysMenuResp
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

func (s *Service) GetByID(ctx context.Context, id uint) (*SysMenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var out *SysMenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		m, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		out = toResp(m)
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysMenuNotFoundDB) {
			return nil, ErrSysMenuNotFound
		}
		return nil, err
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, req CreateSysMenuReq, operatorID uint) (*SysMenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	visible, enabled := true, true
	if req.Visible != nil {
		visible = *req.Visible
	}
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
		CreatedBy: operatorID,
	}
	var out *SysMenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		m, err := s.repo.Create(txCtx, repoReq)
		if err != nil {
			return err
		}
		out = toResp(m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateSysMenuReq, operatorID uint) (*SysMenuResp, error) {
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
		Visible:   req.Visible,
		Enabled:   req.Enabled,
		UpdatedBy: operatorID,
	}
	var out *SysMenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		m, err := s.repo.Update(txCtx, id, repoReq)
		if err != nil {
			return err
		}
		out = toResp(m)
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysMenuNotFoundDB) {
			return nil, ErrSysMenuNotFound
		}
		return nil, err
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, id uint, operatorID uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id, operatorID)
	})
	if err != nil {
		if errors.Is(err, errSysMenuNotFoundDB) {
			return ErrSysMenuNotFound
		}
		return err
	}
	return nil
}

package sysrole

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

func (s *Service) List(ctx context.Context, q ListQuery) ([]*SysRoleResp, int64, error) {
	if s.repo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	page, size := q.Page, q.Size
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}
	var out []*SysRoleResp
	var total int64
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		roles, n, err := s.repo.List(txCtx, q.Keyword, page, size)
		if err != nil {
			return err
		}
		total = n
		out = make([]*SysRoleResp, len(roles))
		for i := range roles {
			out[i] = toResp(&roles[i])
		}
		return nil
	})
	return out, total, err
}

func (s *Service) GetByID(ctx context.Context, id uint) (*SysRoleResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var out *SysRoleResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		role, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		menus, err := s.repo.ListMenus(txCtx, id)
		if err != nil {
			return err
		}
		perms, err := s.repo.ListPermissions(txCtx, id)
		if err != nil {
			return err
		}
		out = toResp(role)
		out.Menus = menus
		out.Permissions = perms
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysRoleNotFoundDB) {
			return nil, ErrSysRoleNotFound
		}
		return nil, err
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, req CreateSysRoleReq, operatorID uint) (*SysRoleResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.Status != nil && *req.Status != 0 && *req.Status != 1 {
		return nil, ErrSysRoleInvalidStatus
	}
	var out *SysRoleResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		status := int8(1)
		if req.Status != nil {
			status = *req.Status
		}
		created, err := s.repo.Create(txCtx, CreateRepoReq{
			OrgID:       req.OrgID,
			Code:        req.Code,
			Name:        req.Name,
			Description: req.Description,
			DataScope:   req.DataScope,
			IsDefault:   req.IsDefault,
			Sort:        req.Sort,
			Status:      status,
			CreatedBy:   operatorID,
		})
		if err != nil {
			return err
		}
		if err := s.repo.AssignMenus(txCtx, created.ID, req.MenuIDs); err != nil {
			return err
		}
		if err := s.repo.AssignPermissions(txCtx, created.ID, req.PermIDs); err != nil {
			return err
		}
		out = toResp(created)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateSysRoleReq, operatorID uint) (*SysRoleResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.Status != nil && *req.Status != 0 && *req.Status != 1 {
		return nil, ErrSysRoleInvalidStatus
	}
	var out *SysRoleResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		updated, err := s.repo.Update(txCtx, id, UpdateRepoReq{
			OrgID:       req.OrgID,
			Code:        req.Code,
			Name:        req.Name,
			Description: req.Description,
			DataScope:   req.DataScope,
			IsDefault:   req.IsDefault,
			Sort:        req.Sort,
			Status:      req.Status,
			UpdatedBy:   operatorID,
		})
		if err != nil {
			return err
		}
		menus, err := s.repo.ListMenus(txCtx, updated.ID)
		if err != nil {
			return err
		}
		perms, err := s.repo.ListPermissions(txCtx, updated.ID)
		if err != nil {
			return err
		}
		out = toResp(updated)
		out.Menus = menus
		out.Permissions = perms
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysRoleNotFoundDB) {
			return nil, ErrSysRoleNotFound
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
		if errors.Is(err, errSysRoleNotFoundDB) {
			return ErrSysRoleNotFound
		}
		return err
	}
	return nil
}

func (s *Service) AssignMenus(ctx context.Context, id uint, menuIDs []uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	return db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		if _, err := s.repo.GetByID(txCtx, id); err != nil {
			return err
		}
		return s.repo.AssignMenus(txCtx, id, menuIDs)
	})
}

func (s *Service) AssignPermissions(ctx context.Context, id uint, permissionIDs []uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	return db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		if _, err := s.repo.GetByID(txCtx, id); err != nil {
			return err
		}
		return s.repo.AssignPermissions(txCtx, id, permissionIDs)
	})
}

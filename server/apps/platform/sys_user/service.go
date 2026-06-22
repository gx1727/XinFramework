package sysuser

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

// Service 是 sys_user 的业务逻辑层。
//
// 关键不变量：
//  1. 所有 DB 操作走 db.RunInPlatformTx（绕过 RLS，sys_users 不启用 RLS）
//  2. 业务校验：code 唯一、status 合法、account 必须存在
//  3. 创建时若给 RoleIDs，会在同一事务里完成 sys_user_roles 授权
type Service struct {
	pool *pgxpool.Pool
	repo Repository
}

func NewService(pool *pgxpool.Pool, repo Repository) *Service {
	return &Service{pool: pool, repo: repo}
}

// List 列出平台用户。
func (s *Service) List(ctx context.Context, q ListQuery) ([]*SysUserResp, int64, error) {
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
	var out []*SysUserResp
	var total int64
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		users, n, err := s.repo.List(txCtx, q.Keyword, page, size)
		if err != nil {
			return err
		}
		total = n
		out = make([]*SysUserResp, len(users))
		for i := range users {
			out[i] = toResp(&users[i])
		}
		return nil
	})
	return out, total, err
}

// GetByID 按 ID 取一个平台用户，并附上角色列表。
func (s *Service) GetByID(ctx context.Context, id uint) (*SysUserResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var out *SysUserResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		u, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		roles, err := s.repo.ListRoles(txCtx, u.ID)
		if err != nil {
			return err
		}
		u.Roles = roles
		out = toResp(u)
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysUserNotFoundDB) {
			return nil, ErrSysUserNotFound
		}
		return nil, err
	}
	return out, nil
}

// Create 新建平台用户。可选地一次性绑定角色。
func (s *Service) Create(ctx context.Context, req CreateSysUserReq, operatorID uint) (*SysUserResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.AccountID == 0 {
		return nil, ErrSysUserAccountRequired
	}
	if req.Status != nil && *req.Status != 0 && *req.Status != 1 {
		return nil, ErrSysUserInvalidStatus
	}

	var out *SysUserResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		// 唯一性预检：同 account_id 已存在则报错
		if existing, err := s.repo.GetByAccountID(txCtx, req.AccountID); err == nil && existing != nil {
			return ErrSysUserAlreadyExists
		} else if err != nil && !errors.Is(err, errSysUserNotFoundDB) {
			return err
		}

		status := int8(1)
		if req.Status != nil {
			status = *req.Status
		}

		created, err := s.repo.Create(txCtx, CreateRepoReq{
			AccountID: req.AccountID,
			Code:      req.Code,
			OrgID:     req.OrgID,
			RealName:  req.RealName,
			Nickname:  req.Nickname,
			Avatar:    req.Avatar,
			Status:    status,
			CreatedBy: operatorID,
		})
		if err != nil {
			return err
		}
		for _, rid := range req.RoleIDs {
			if err := s.repo.GrantRole(txCtx, created.ID, rid); err != nil {
				return err
			}
		}
		if len(req.RoleIDs) > 0 {
			roles, err := s.repo.ListRoles(txCtx, created.ID)
			if err != nil {
				return err
			}
			created.Roles = roles
		}
		out = toResp(created)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Update 修改平台用户基本信息（不包含角色）。
func (s *Service) Update(ctx context.Context, id uint, req UpdateSysUserReq, operatorID uint) (*SysUserResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.Status != nil && *req.Status != 0 && *req.Status != 1 {
		return nil, ErrSysUserInvalidStatus
	}
	var out *SysUserResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		updated, err := s.repo.Update(txCtx, id, UpdateRepoReq{
			Code:      req.Code,
			OrgID:     req.OrgID,
			RealName:  req.RealName,
			Nickname:  req.Nickname,
			Avatar:    req.Avatar,
			Status:    req.Status,
			UpdatedBy: operatorID,
		})
		if err != nil {
			return err
		}
		roles, err := s.repo.ListRoles(txCtx, updated.ID)
		if err != nil {
			return err
		}
		updated.Roles = roles
		out = toResp(updated)
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysUserNotFoundDB) {
			return nil, ErrSysUserNotFound
		}
		return nil, err
	}
	return out, nil
}

// UpdateStatus 单独改状态。
func (s *Service) UpdateStatus(ctx context.Context, id uint, status int8, operatorID uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	if status != 0 && status != 1 {
		return ErrSysUserInvalidStatus
	}
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		return s.repo.UpdateStatus(txCtx, id, status, operatorID)
	})
	if err != nil {
		if errors.Is(err, errSysUserNotFoundDB) {
			return ErrSysUserNotFound
		}
		return err
	}
	return nil
}

// Delete 软删除。
func (s *Service) Delete(ctx context.Context, id uint, operatorID uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id, operatorID)
	})
	if err != nil {
		if errors.Is(err, errSysUserNotFoundDB) {
			return ErrSysUserNotFound
		}
		return err
	}
	return nil
}

// AssignRoles 全量替换用户角色集合。
func (s *Service) AssignRoles(ctx context.Context, id uint, roleIDs []uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	return db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		// 先确认 user 存在
		if _, err := s.repo.GetByID(txCtx, id); err != nil {
			return err
		}
		// 全量重写：先 revoke 全部，再 grant 目标集合。
		// 为简单起见走 DELETE + INSERT：
		if _, err := txCtxExec(txCtx, s.pool, `DELETE FROM sys_user_roles WHERE user_id = $1`, id); err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := s.repo.GrantRole(txCtx, id, rid); err != nil {
				return err
			}
		}
		return nil
	})
}

// txCtxExec 是为了让 AssignRoles 内部能在不通过 repo 的情况下
// 跑一条简单 SQL。service 通过 ctx 拿到 Querier。
func txCtxExec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error) {
	q, err := db.GetQuerier(ctx, pool)
	if err != nil {
		return 0, err
	}
	cmd, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

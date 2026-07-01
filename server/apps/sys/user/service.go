package sysuser

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/db"
)

// Service 是 sys_user 的业务逻辑层。
//
// 关键不变量：
//  1. 所有 DB 操作走 db.RunInSysTx（绕过 RLS，sys_users 不启用 RLS）
//  2. 业务校验：code 唯一、status 合法、account 必须存在
//  3. 创建时若给 RoleIDs，会在同一事务里完成 sys_user_roles 授权
//  4. 模式 2（一并建账号）：AccountID == 0 时由本服务在同一事务内
//     通过 accountRepo.Create 同步建 accounts 行，确保原子性。
type Service struct {
	pool        *pgxpool.Pool
	repo        Repository
	accountRepo auth.AccountRepository // 可为 nil（未注入时禁用模式 2）
}

func NewService(pool *pgxpool.Pool, repo Repository, accountRepo auth.AccountRepository) *Service {
	return &Service{pool: pool, repo: repo, accountRepo: accountRepo}
}

// List 列出 sys 用户，并为每个用户填上已分配的 sys 角色列表。
//
// ⭐ 修复：原实现只返了 sys_users 本身，未填 Roles——前端 SysUsers 页"分配 sys 角色" dialog
// 打开时拿不到已选中的角色（user.roles 为 undefined），全部选项都未勾。
// 现与 Service.GetByID 保持一致，在事务内对每个用户调一次 ListRoles。
//
// 性能：N+1 查询（len(users) 次）。考虑在 page=200 、200 个用户的场景下，总计 201 次
// 简单查询（主键等值 IN，已走索引）。后续如要优化可改 ListRolesByUserIDs([]uint) 批查。
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
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		users, n, err := s.repo.List(txCtx, q.Keyword, page, size)
		if err != nil {
			return err
		}
		total = n
		out = make([]*SysUserResp, len(users))
		for i := range users {
			roles, err := s.repo.ListRoles(txCtx, users[i].ID)
			if err != nil {
				return err
			}
			users[i].Roles = roles
			out[i] = toResp(&users[i])
		}
		return nil
	})
	return out, total, err
}

// GetByID 按 ID 取一个 sys 用户，并附上角色列表。
func (s *Service) GetByID(ctx context.Context, id uint) (*SysUserResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var out *SysUserResp
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
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

// Create 新建 sys 用户。可选地一次性绑定角色。
//
// 支持两种模式（详见 CreateSysUserReq）：
//   - AccountID > 0：绑定已有账号。
//   - AccountID == 0：在一同事务内新建账号并绑定。
func (s *Service) Create(ctx context.Context, req CreateSysUserReq, operatorID uint) (*SysUserResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.Status != nil && *req.Status != 0 && *req.Status != 1 {
		return nil, ErrSysUserInvalidStatus
	}

	var out *SysUserResp
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		// ---- 模式 2：一并新建可登录账号 ----
		if req.AccountID == 0 {
			if s.accountRepo == nil {
				return ErrSysUserAccountRepoMissing
			}
			if err := s.createAccountInline(txCtx, &req); err != nil {
				return err
			}
		}

		// 唯一性预检：同 account_id 已存在则报错
		if existing, err := s.repo.GetByAccountID(txCtx, req.AccountID); err == nil && existing != nil {
			return ErrSysUserAlreadyExists
		} else if err != nil && !errors.Is(err, errSysUserNotFoundDB) {
			return err
		}

		// code 留空时按 "u<account_id>" 自动生成
		if req.Code == "" {
			req.Code = fmt.Sprintf("u%d", req.AccountID)
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

// createAccountInline 在当前事务里创建 accounts 行并回填 req.AccountID。
//
// 语义：
//   - phone / password 必填（由 binding tag 负责，service 也再确认一次作为纵深防御）
//   - phone 唯一：先 GetByPhone 预检，命中则返回 ErrSysUserPhoneExists
//   - email 选填：非空时同样预检 ErrSysUserEmailExists
//   - password 经 framework/pkg/auth.HashPassword 转 Argon2id 后入库
//   - username 留空时默认同 phone
//
// 必须在 RunInSysTx 回调内调用。出错后外层事务会一起回滚。
func (s *Service) createAccountInline(ctx context.Context, req *CreateSysUserReq) error {
	if req.Phone == "" {
		return ErrSysUserPhoneRequired
	}
	if len(req.Password) < 6 || len(req.Password) > 32 {
		return ErrSysUserPasswordInvalid
	}

	if existing, err := s.accountRepo.GetByPhone(ctx, req.Phone); err == nil && existing != nil {
		return ErrSysUserPhoneExists
	} else if err != nil && !errors.Is(err, auth.ErrAccountNotFound) {
		return err
	}

	if req.Email != "" {
		if existing, err := s.accountRepo.GetByEmail(ctx, req.Email); err == nil && existing != nil {
			return ErrSysUserEmailExists
		} else if err != nil && !errors.Is(err, auth.ErrAccountNotFound) {
			return err
		}
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	username := req.Username
	if username == "" {
		username = req.Phone
	}

	acc, err := s.accountRepo.Create(ctx, username, req.Phone, req.Email, req.RealName, hash)
	if err != nil {
		return err
	}
	req.AccountID = acc.ID
	return nil
}

// Update 修改 sys 用户基本信息（不包含角色）。
func (s *Service) Update(ctx context.Context, id uint, req UpdateSysUserReq, operatorID uint) (*SysUserResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.Status != nil && *req.Status != 0 && *req.Status != 1 {
		return nil, ErrSysUserInvalidStatus
	}
	var out *SysUserResp
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
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
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
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
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
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
	return db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
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

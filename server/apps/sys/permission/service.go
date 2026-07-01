package syspermission

import (
	"context"
	"errors"
	"strings"

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

// permissionCodeValid 校验权限码格式。
// 规则（0024+）：
//   - 纯字符串（如 changepwd）：菜单无关资源，无 action 维度。允许。
//   - "resource:action" 或 "resource:*"：菜单相关资源，必须含且仅含一个 ":"，前后非空。
//   - 多个 ":" 拒绝。
func permissionCodeValid(code string) bool {
	if code == "" {
		return false
	}
	count := strings.Count(code, ":")
	switch count {
	case 0:
		return true
	case 1:
		idx := strings.Index(code, ":")
		return idx > 0 && idx < len(code)-1
	default:
		return false
	}
}

func (s *Service) List(ctx context.Context, q ListQuery) ([]*SysPermissionResp, int64, error) {
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
	var out []*SysPermissionResp
	var total int64
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		perms, n, err := s.repo.List(txCtx, q.MenuID, q.Keyword, page, size)
		if err != nil {
			return err
		}
		total = n
		out = make([]*SysPermissionResp, len(perms))
		for i := range perms {
			out[i] = toResp(&perms[i])
		}
		return nil
	})
	return out, total, err
}

func (s *Service) GetByID(ctx context.Context, id uint) (*SysPermissionResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var out *SysPermissionResp
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		p, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		out = toResp(p)
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysPermissionNotFoundDB) {
			return nil, ErrSysPermissionNotFound
		}
		return nil, err
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, req CreateSysPermissionReq, operatorID uint) (*SysPermissionResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if !permissionCodeValid(req.Code) {
		return nil, ErrSysPermissionInvalidCode
	}
	repoReq := CreateRepoReq{
		MenuID:      req.MenuID,
		Code:        req.Code,
		Name:        req.Name,
		Action:      req.Action,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      1,
		CreatedBy:   operatorID,
	}
	if req.Status != nil {
		repoReq.Status = *req.Status
	}
	var out *SysPermissionResp
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		// 唯一性预检：GetByCode 返回空切片时无冲突
		existing, err := s.repo.GetByCode(txCtx, req.Code)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return ErrSysPermissionCodeExists
		}
		p, err := s.repo.Create(txCtx, repoReq)
		if err != nil {
			// 兜底：触发 uk_sys_permissions_code 唯一约束时报业务错误
			if isUniqueViolation(err) {
				return ErrSysPermissionCodeExists
			}
			return err
		}
		out = toResp(p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateSysPermissionReq, operatorID uint) (*SysPermissionResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	if req.Code != nil && !permissionCodeValid(*req.Code) {
		return nil, ErrSysPermissionInvalidCode
	}
	repoReq := UpdateRepoReq{
		MenuID:      req.MenuID,
		Code:        req.Code,
		Name:        req.Name,
		Action:      req.Action,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
		UpdatedBy:   operatorID,
	}
	var out *SysPermissionResp
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		p, err := s.repo.Update(txCtx, id, repoReq)
		if err != nil {
			return err
		}
		out = toResp(p)
		return nil
	})
	if err != nil {
		if errors.Is(err, errSysPermissionNotFoundDB) {
			return nil, ErrSysPermissionNotFound
		}
		return nil, err
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, id uint, operatorID uint) error {
	if s.repo == nil {
		return ErrBackendUnavailable
	}
	err := db.RunInSysTx(ctx, s.pool, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id, operatorID)
	})
	if err != nil {
		if errors.Is(err, errSysPermissionNotFoundDB) {
			return ErrSysPermissionNotFound
		}
		return err
	}
	return nil
}

// isUniqueViolation 判断 pgx 错误是否是 23505（unique_violation）。
// service 层不直接依赖 pgconn 内部类型，用字符串包含 SQLSTATE 即可。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") || strings.Contains(msg, "duplicate key")
}

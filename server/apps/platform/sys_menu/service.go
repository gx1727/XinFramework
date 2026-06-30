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

// Tree 返回平台菜单树。super_admin 全量；其他 platform 角色仅返回其角色被分配的菜单。
//
// callerAccountID 来自 JWT Claims.UserID（平台用户 = account_id，参见
// apps/boot/auth.PlatformLogin）。callerRoles 来自 XinContext.PlatformRoles。
//
// 为什么把过滤逻辑放在 service 而不是 handler：过滤在 RunInPlatformTx 事务里走，
// 与现有 GetAll 的事务路径保持一致；handler 只负责从 gin ctx 拼出参数。
func (s *Service) Tree(ctx context.Context, callerAccountID uint, callerRoles []string) ([]*SysMenuResp, error) {
	if s.repo == nil {
		return nil, ErrBackendUnavailable
	}
	var tree []*SysMenuResp
	err := db.RunInPlatformTx(ctx, s.pool, func(txCtx context.Context) error {
		var menus []Menu
		var err error
		if isSuperAdmin(callerRoles) {
			menus, err = s.repo.GetAll(txCtx)
		} else {
			// 零角色送进来会被中间件拦住，这里加个安全勾兑：
			// 万一被直接调用不经过中间件，返回空树而不是泄漏全量。
			menus, err = s.repo.ListByUserRoles(txCtx, callerAccountID, callerRoles)
		}
		if err != nil {
			return err
		}
		tree = buildTree(menus)
		return nil
	})
	return tree, err
}

// isSuperAdmin 脱耦于 jwt.PlatformRoleSuperAdmin 常量，避免 sysmenu 包对 jwt 包的直接依赖。
func isSuperAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "super_admin" {
			return true
		}
	}
	return false
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

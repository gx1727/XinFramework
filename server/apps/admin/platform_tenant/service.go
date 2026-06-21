package platformtenant

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	pool       *pgxpool.Pool
	tenantRepo TenantRepository
}

func NewService(pool *pgxpool.Pool, repo TenantRepository) *Service {
	return &Service{pool: pool, tenantRepo: repo}
}

// GetByID 平台级操作：走 RunInPlatformTx，绕过 RLS。
// 即使 token 内带 tenant_id 也不影响——bypass_rls=on 后任何 tenants 行都能读到。
func (s *Service) GetByID(ctx context.Context, id uint) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var t *Tenant
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		t, err = s.tenantRepo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	resp := toResp(t)
	return &resp, nil
}

// Create 平台级操作：创建租户 + 首装（root org / admin role / starter dicts / admin user）。
//
// 全程单一 RunInPlatformTx 事务：tenant INSERT → firstInstall → commit。
// 任一环节失败回滚，租户不留半成品。
//
// 首装内容：
//   - root 组织（每个租户必须有，作为后续组织祖先）
//   - admin role + 绑定超级资源 *:*（保证全权限）
//   - admin user（若 req.AdminAccountID 提供且账号存在）
//   - starter dicts: gender / user_status / education
//   - tenant_user_seq 初始化（用于 user_code 自增）
func (s *Service) Create(ctx context.Context, req CreateTenantReq) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var (
		t   *Tenant
		rep *FirstInstallReport
	)
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		t, err = s.tenantRepo.Create(ctx, req.Code, req.Name, req.Contact, req.Phone, req.Email)
		if err != nil {
			return err
		}
		// 首装：单一事务内继续，失败全回滚
		var adminAccountID uint
		if req.AdminAccountID != nil {
			adminAccountID = *req.AdminAccountID
		}
		rep, err = firstInstall(ctx, s.pool, t.ID, adminAccountID)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	// 审计：tenant.create + 首装明细 —— 高敏操作必须留痕。
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:create",
		TableName: "tenants",
		RecordID:  t.ID,
		NewData: map[string]any{
			"id":      t.ID,
			"code":    t.Code,
			"name":    t.Name,
			"status":  t.Status,
			"contact": t.Contact,
			"phone":   t.Phone,
			"email":   t.Email,
			"first_install": map[string]any{
				"template":                    TemplateTenantCode,
				"root_org_id":                 rep.RootOrgID,
				"admin_role_id":               rep.AdminRoleID,
				"admin_user_id":               rep.AdminUserID,
				"menu_count":                  rep.MenuCount,
				"resource_count":              rep.ResourceCount,
				"dict_count":                  rep.DictCount,
				"dict_item_count":             rep.DictItemCount,
				"tenant_user_seq_initialized": rep.TenantUserSeqInit,
				"warnings":                    rep.WarnMessages,
			},
		},
	})
	resp := toResp(t)
	return &resp, nil
}

// Update 平台级操作：档案字段更新。OldData 在改前抓快照便于审计 diff。
func (s *Service) Update(ctx context.Context, id uint, req UpdateTenantReq) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var before *Tenant
	var after *Tenant
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		before, err = s.tenantRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		after, err = s.tenantRepo.Update(ctx, id, req.Name, req.Contact, req.Phone, req.Email,
			req.Province, req.City, req.Area, req.Address)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:update",
		TableName: "tenants",
		RecordID:  id,
		OldData: map[string]any{
			"name":     before.Name,
			"contact":  before.Contact,
			"phone":    before.Phone,
			"email":    before.Email,
			"province": before.Province,
			"city":     before.City,
			"area":     before.Area,
			"address":  before.Address,
		},
		NewData: map[string]any{
			"name":     after.Name,
			"contact":  after.Contact,
			"phone":    after.Phone,
			"email":    after.Email,
			"province": after.Province,
			"city":     after.City,
			"area":     after.Area,
			"address":  after.Address,
		},
	})
	resp := toResp(after)
	return &resp, nil
}

// UpdateStatus 单独的状态切换（启用 / 禁用）。比通用 Update 更便于权限细分与审计。
func (s *Service) UpdateStatus(ctx context.Context, id uint, status int16) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var before *Tenant
	var after *Tenant
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		before, err = s.tenantRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		after, err = s.tenantRepo.UpdateStatus(ctx, id, status)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:status_change",
		TableName: "tenants",
		RecordID:  id,
		OldData:   map[string]any{"status": before.Status},
		NewData:   map[string]any{"status": after.Status},
	})
	resp := toResp(after)
	return &resp, nil
}

// Delete 平台级操作：前置校验 + 软删 + 审计三步走。
//
// 前置校验：租户下存在 is_deleted=FALSE 的 users 时禁止软删，
// 防止留下"幽灵租户"（status=1、is_deleted=FALSE、但已被标记删除）。
// 必须先把所有用户迁出 / 软删，再走 tenant.Delete。
func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.tenantRepo == nil {
		return ErrBackendUnavailable
	}
	var tenantCode string
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		// 1) 拿租户快照（用于审计 + 校验）
		t, err := s.tenantRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		tenantCode = t.Code

		// 2) 前置校验：未软删用户数
		n, err := s.tenantRepo.CountActiveUsers(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w (用户数=%d)", ErrTenantHasUsers, n)
		}

		// 3) 软删
		return s.tenantRepo.Delete(ctx, id)
	})
	if err != nil {
		return mapRepoError(err)
	}

	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:delete",
		TableName: "tenants",
		RecordID:  id,
		OldData:   map[string]any{"code": tenantCode, "is_deleted": false},
		NewData:   map[string]any{"code": tenantCode, "is_deleted": true},
	})
	return nil
}

// List 平台级操作：所有租户可见。
// 注意：传 ctx 里的 tenant_id（如果有）也不会影响 List——bypass_rls 屏蔽 RLS。
func (s *Service) List(ctx context.Context, req ListTenantReq) ([]TenantResp, int64, error) {
	if s.tenantRepo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	var list []Tenant
	var total int64
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		list, total, err = s.tenantRepo.List(ctx, req.Keyword, req.Status, req.Page, req.Size)
		return err
	})
	if err != nil {
		return nil, 0, mapRepoError(err)
	}
	resps := make([]TenantResp, len(list))
	for i := range list {
		resps[i] = toResp(&list[i])
	}
	return resps, total, nil
}

// PurgeResult 硬删返回：保留删除明细便于审计与回查。
type PurgeResult struct {
	TenantID uint
	Code     string
	// Tables 每张表实际删除的行数。审计时直接写入 db_logs.new_data。
	Tables map[string]int64
}

// Purge 硬删租户及其全部业务数据。**不可逆操作**，仅 super_admin 可调。
//
// 流程：
//  1. 前置校验：租户必须已软删（is_deleted=TRUE）——避免误删正在使用的租户。
//  2. 前置校验：租户下不存在 is_deleted=FALSE 的 users（防止留下 FK 孤儿）。
//  3. RunInPlatformTx 内：
//     a. 抓快照（code 用于审计）
//     b. PurgeTenantData 按依赖顺序硬删 17 张 tenant_id-bearing 表
//     c. HardDelete 删 tenants 表本身
//  4. 审计：写 db_logs，new_data 含每张表删除行数。
//
// 失败回滚：事务原子性保证——若任一 DELETE 失败，所有改动回滚，租户仍为软删状态，
// 排查后可重试。
func (s *Service) Purge(ctx context.Context, id uint) (*PurgeResult, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var result PurgeResult
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		// 1) 抓快照 + 前置校验：必须已软删
		t, err := s.tenantRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if !t.IsDeleted {
			return ErrTenantPurgeNotAllowed
		}
		result.TenantID = t.ID
		result.Code = t.Code

		// 2) 前置校验：无活跃用户
		n, err := s.tenantRepo.CountActiveUsers(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w (用户数=%d)", ErrTenantHasUsers, n)
		}

		// 3a) 硬删所有 tenant_id-bearing 表
		tables, err := s.tenantRepo.PurgeTenantData(ctx, id)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrTenantPurgeFailed, err)
		}
		result.Tables = tables

		// 3b) 硬删 tenants 表本身
		if err := s.tenantRepo.HardDelete(ctx, id); err != nil {
			return fmt.Errorf("%w: %v", ErrTenantPurgeFailed, err)
		}
		return nil
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	// 4) 审计：写明每张表删了多少行——事后无据可查，这是最后证据。
	totalDeleted := int64(0)
	for _, n := range result.Tables {
		totalDeleted += n
	}
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:purge",
		TableName: "tenants",
		RecordID:  id,
		OldData: map[string]any{
			"code":            result.Code,
			"is_deleted":      true,
			"tables_to_purge": len(result.Tables),
		},
		NewData: map[string]any{
			"code":          result.Code,
			"tables_purged": len(result.Tables),
			"total_rows":    totalDeleted,
			"tables":        result.Tables,
		},
	})
	return &result, nil
}

package tenants

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/xincontext"
)

type Service struct {
	pool       *pgxpool.Pool
	tenantRepo TenantRepository
	jwtCfg     *config.JWTConfig
}

func NewService(pool *pgxpool.Pool, repo TenantRepository, jwtCfg *config.JWTConfig) *Service {
	return &Service{pool: pool, tenantRepo: repo, jwtCfg: jwtCfg}
}

// GetByID sys 级操作：赀RunInSysTx，绕迀RLS
// 即使 token 内带 tenant_id 也不影响——bypass_rls=on 后任佀tenants 行都能读到
func (s *Service) GetByID(ctx context.Context, id uint) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var t *Tenant
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
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

// Create sys 级操作：创建租户 + 首装（root org / admin role / starter dicts / admin user）
//
// 全程单一 RunInSysTx 事务：tenant INSERT ↀfirstInstall ↀcommit
// 任一环节失败回滚，租户不留半成品
//
// admin_account_id 解析规则：
//   - req.AdminAccountID != nil && *req.AdminAccountID > 0：用 req 提供的账号（精确绑定）
//   - 否则：fallback 用当前 super_admin 的 account_id（= Context.UserID，sys 域登录时即 account_id）
//   - 这样 super_admin 创建的每个租户都自动有 super_admin 自己的 admin 身份，
//     后续可直接用 POST /tenants/:id/impersonate 切进去验收。
//   - 极端情况（super_admin 缺失 / context 不可读）：fallback 为 0，
//     installAdminUser 走 warn 跳过分支；后续可手动补 admin user。
//
// 首装内容：
//   - root 组织（每个租户必须有，作为后续组织祖先）
//   - admin role + 绑定超级资源 *:*（保证全权限：
//   - admin user（若 adminAccountID > 0 且账号存在）
//   - starter dicts: gender / user_status / education
//   - tenant_user_seq 初始化（用于 user_code 自增：
func (s *Service) Create(ctx context.Context, req CreateTenantReq) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}

	// 取 actor（super_admin 的 account_id）—— 用于 admin_account_id fallback
	var actorAccountID uint
	if xc, ok := xincontext.XinContextFrom(ctx); ok {
		actorAccountID = xc.UserID
	}

	var (
		t   *Tenant
		rep *FirstInstallReport
	)
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		t, err = s.tenantRepo.Create(ctx, req.Code, req.Name, req.Contact, req.Phone, req.Email)
		if err != nil {
			return err
		}
		// 首装：单一事务内继续，失败全回滀
		var adminAccountID uint
		if req.AdminAccountID != nil && *req.AdminAccountID > 0 {
			adminAccountID = *req.AdminAccountID
		} else {
			adminAccountID = actorAccountID // fallback：super_admin 自身
		}
		rep, err = firstInstall(ctx, s.pool, t.ID, adminAccountID)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	// 审计：tenant.create + 首装明细 — 高敏操作必须留痕
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:create",
		TableName: "tenants",
		RecordID:  t.ID,
		TenantID:  t.ID,
		UserID:    actorAccountID,
		NewData: map[string]any{
			"id":               t.ID,
			"code":             t.Code,
			"name":             t.Name,
			"status":           t.Status,
			"contact":          t.Contact,
			"phone":            t.Phone,
			"email":            t.Email,
			"actor_account_id": actorAccountID,
			"admin_account_id": resolveAdminAccountID(req.AdminAccountID, actorAccountID),
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

// resolveAdminAccountID audit 中显式记录最终使用的 admin_account_id 来源。
func resolveAdminAccountID(reqID *uint, actor uint) uint {
	if reqID != nil && *reqID > 0 {
		return *reqID
	}
	return actor
}

// Update sys 级操作：档案字段更新。OldData 在改前抓快照便于审计 diff
func (s *Service) Update(ctx context.Context, id uint, req UpdateTenantReq) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var before *Tenant
	var after *Tenant
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
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

// UpdateStatus 单独的状态切换（启用 / 禁用）。比通用 Update 更便于权限细分与审计
func (s *Service) UpdateStatus(ctx context.Context, id uint, status int16) (*TenantResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var before *Tenant
	var after *Tenant
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
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

// Delete sys 级操作：前置校验 + 软删 + 审计三步走
//
// 前置校验：租户下存在 is_deleted=FALSE 皀users 时禁止软删，
// 防止留下"幽灵租户"（status=1、is_deleted=FALSE、但已被标记删除）
// 必须先把所有用户迁净/ 软删，再赀tenant.Delete
func (s *Service) Delete(ctx context.Context, id uint) error {
	if s.tenantRepo == nil {
		return ErrBackendUnavailable
	}
	var tenantCode string
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
		// 1) 拿租户快照（用于审计 + 校验：
		t, err := s.tenantRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		tenantCode = t.Code

		// 2) 前置校验：未软删用户敀
		n, err := s.tenantRepo.CountActiveUsers(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w (用户敀%d)", ErrTenantHasUsers, n)
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

// List sys 级操作：所有租户可见
// 注意：传 ctx 里的 tenant_id（如果有）也不会影响 List——bypass_rls 屏蔽 RLS
func (s *Service) List(ctx context.Context, req ListTenantReq) ([]TenantResp, int64, error) {
	if s.tenantRepo == nil {
		return nil, 0, ErrBackendUnavailable
	}
	var list []Tenant
	var total int64
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
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

// PurgeResult 硬删返回：保留删除明细便于审计与回查
type PurgeResult struct {
	TenantID uint
	Code     string
	// Tables 每张表实际删除的行数。审计时直接写入 db_logs.new_data
	Tables map[string]int64
}

// purgeTables 硬删租户数据的顺序。按"先删子表、再删父血的依赖关系排列，
// 避免 FK 约束冲突。每个表名对庀migrations/framework.sql 中带 tenant_id 的表
//
// 注意：即使没有显开FK 约束（migrations 里多数未声明 FK），也按业务依赖排序：
//   - usage_records / attachments ：纯记录表，无业务依赀
//   - subscriptions：依赀plans（但 plans 是平台级，不删）
//   - role_data_scopes / user_roles / role_menus / role_resources：依赀roles
//   - dict_items：依赀dicts
//   - routes / resources / menus：依赀organizations / 吀role_resource 关系
//   - organizations / roles / users / dicts：核心实佀
//   - tenant_user_seq：独立序列表
//
// 表名来自 migrations 文件常量列表，不是用户输入——安全
// 赀$1 参数化防注入（虽焀table 不会被替换）
var purgeTables = []string{
	"usage_records",
	"attachments",
	"subscriptions",
	"role_data_scopes",
	"user_roles",
	"role_menus",
	"role_resources",
	"dict_items",
	"routes",
	"resources",
	"menus",
	"organizations",
	"roles",
	"users",
	"dicts",
	"tenant_user_seq",
}

// Purge 硬删租户及其全部业务数据　*不可逆操佀*，仅 super_admin 可调
//
// 流程：
//  1. 前置校验：租户必须已软删（is_deleted=TRUE）——避免误删正在使用的租户
//  2. 前置校验：租户下不存圀is_deleted=FALSE 皀users（防止留一FK 孤儿）
//  3. RunInSysTx 内：直接甀s.pool 跀SQL（跨多表的复合业务操作）
//     a. 抓快照（code 用于审计：
//     b. 挀purgeTables 顺序 DELETE tenant_id-bearing 血
//     c. DELETE tenants 表本躀
//  4. 审计：写 db_logs，new_data 含每张表删除行数
//
// 失败回滚：事务原子性保证——若任一 DELETE 失败，所有改动回滚，租户仍为软删状态，
// 排查后可重试
//
// Purge 是跨多表的业务级操作（不是单一表的 CRUD），因此直接甀s.pool 跀SQL：
// 不下沉到 Repository
func (s *Service) Purge(ctx context.Context, id uint) (*PurgeResult, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	var result PurgeResult
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx, s.pool)
		if err != nil {
			return err
		}

		// 1) 抓快煀+ 前置校验：必须已软删
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
			return fmt.Errorf("%w (用户敀%d)", ErrTenantHasUsers, n)
		}

		// 3) 硬删所最tenant_id-bearing 血+ tenants 本身
		result.Tables = make(map[string]int64, len(purgeTables))
		for _, table := range purgeTables {
			tag, err := q.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = $1`, table), id)
			if err != nil {
				return fmt.Errorf("%w: purge %s: %v", ErrTenantPurgeFailed, table, err)
			}
			result.Tables[table] = tag.RowsAffected()
		}
		tag, err := q.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
		if err != nil {
			return fmt.Errorf("%w: hard delete tenants: %v", ErrTenantPurgeFailed, err)
		}
		if tag.RowsAffected() == 0 {
			return ErrTenantNotFound
		}
		return nil
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	// 4) 审计：写明每张表删了多少行——事后无据可查，这是最后证据
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

// Impersonate sys 管理员（super_admin）模拟登录到指定租户。
//
// 业务动机：super_admin 创建新租户后，新租户通常没有绑定任何账号，
// 无法用现有 /auth/select-tenant 切进去验证。本端点签发一个
// 临时的"模拟 token"，让 super_admin 以租户 admin 角色身份进入，
// 走完整租户 RBAC（不享受 sys 域短路），操作可审计。
//
// 流程（事务保证原子性）：
//  1. 校验租户存在且 status=1
//  2. 找租户的 admin user（first_install 创建的 admin role 绑定用户）
//  3. 在 auth_sessions 写入新会话（account_id = 当前 super_admin account）
//  4. 签 access + refresh token，把 ImpersonatedBy/ImpersonationSID 写入 claims
//  5. 写 audit：tenant:impersonate_start，含 actor/tenant/admin user 关联
//
// 退出模拟：前端保存原 sys refresh_token，调 /auth/refresh（不传 tenant_id）
// 即可用原 sys 会话恢复 sys token。
//
// RunInSysTx 是必须的：tenant_users / tenant_user_roles / tenant_roles
// 都有 RLS，bypass_rls=on 才能跨租户查询。
func (s *Service) Impersonate(ctx context.Context, tenantID uint) (*ImpersonateResp, error) {
	if s.tenantRepo == nil {
		return nil, ErrBackendUnavailable
	}
	if s.jwtCfg == nil {
		return nil, fmt.Errorf("%w: jwt config missing", ErrTenantImpersonateFailed)
	}

	// 1) 取 actor（当前 super_admin account_id）—— sys 域登录时 JWT.UserID = account_id
	var actorAccountID uint
	var originalSID string
	if xc, ok := xincontext.XinContextFrom(ctx); ok {
		actorAccountID = xc.UserID
		originalSID = string(xc.SessionID)
	}
	if actorAccountID == 0 {
		return nil, fmt.Errorf("%w: actor account_id missing (must be sys login)", ErrTenantImpersonateFailed)
	}

	var (
		resp     ImpersonateResp
		adminUsr *AdminUser
		tenant   *Tenant
	)
	err := db.RunInSysTx(ctx, s.pool, func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx, s.pool)
		if err != nil {
			return err
		}

		// a) 校验租户存在
		t, err := s.tenantRepo.GetByID(ctx, tenantID)
		if err != nil {
			return err
		}
		if t.Status != 1 {
			return ErrTenantDisabled
		}
		tenant = t

		// b) 找 admin role 绑定的用户
		adminUsr, err = s.tenantRepo.FindAdminUser(ctx, tenantID)
		if err != nil {
			return err
		}

		// c) 写 auth_session（按 account 维度；session_id = uuid）
		sid := uuid.NewString()
		expireSec := s.jwtCfg.Expire
		if _, err := q.Exec(ctx, `
			INSERT INTO auth_sessions (account_id, token, ip, user_agent, expires_at)
			VALUES ($1, $2, '', 'impersonation', NOW() + make_interval(secs => $3))`,
			actorAccountID, sid, expireSec); err != nil {
			return fmt.Errorf("create impersonation session: %w", err)
		}

		// d) 签 access + refresh token
		access, err := jwtpkg.GenerateImpersonation(s.jwtCfg,
			adminUsr.ID, tenantID, "admin",
			sid, actorAccountID, originalSID,
			jwtpkg.TokenTypeAccess)
		if err != nil {
			return fmt.Errorf("sign access token: %w", err)
		}
		refresh, err := jwtpkg.GenerateImpersonation(s.jwtCfg,
			adminUsr.ID, tenantID, "admin",
			sid, actorAccountID, originalSID,
			jwtpkg.TokenTypeRefresh)
		if err != nil {
			return fmt.Errorf("sign refresh token: %w", err)
		}

		resp = ImpersonateResp{
			Scope:              ImpersonateScopeTenant,
			Token:              access,
			RefreshToken:       refresh,
			ExpiresIn:          s.jwtCfg.Expire,
			TenantID:           tenantID,
			TenantName:         tenant.Name,
			ImpersonatedUserID: adminUsr.ID,
			ImpersonatedBy:     actorAccountID,
			ImpersonationSID:   originalSID,
		}
		return nil
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	// 审计（事务外写）：写明 actor、原 sys SID、目标租户、admin user
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "tenant:impersonate_start",
		TableName: "tenants",
		RecordID:  tenantID,
		TenantID:  tenantID,
		UserID:    actorAccountID,
		NewData: map[string]any{
			"actor_account_id":     actorAccountID,
			"original_session_id":  originalSID,
			"target_tenant_id":     tenantID,
			"target_tenant_code":   tenant.Code,
			"target_tenant_name":   tenant.Name,
			"impersonated_user_id": adminUsr.ID,
			"impersonated_user":    adminUsr.RealName,
			"role":                 "admin",
		},
	})
	return &resp, nil
}

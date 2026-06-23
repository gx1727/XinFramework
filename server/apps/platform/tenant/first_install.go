package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// TemplateTenantCode 模板租户的特殊 code。所有 first_install 流程从此租户复制
// menus / resources / dicts / dict_items / config_categories / config_items。
//
// 历史方案：曾有一个 __template__ (status=0) 作为克隆源 + default (status=1) 作为 admin 居住地。
// 现已合并为单一 bootstrap 租户同时承担两个角色（admin 居住 + 新租户克隆源）。
//   - status=1（激活），所以 admin 能登录看到所有模板数据
//   - new_install.go 显式只克隆数据表（menus/resources/dicts/config_categories/config_items），
//     不克隆 users / organizations / user_roles / role_data_scopes / tenant_user_seq
//   - 这些"租户级独有"的数据每次首装独立创建
const TemplateTenantCode = "bootstrap"

// FirstInstallReport 记录首装每一步的结果，供 service 写入 audit。
type FirstInstallReport struct {
	RootOrgID         uint
	AdminUserID       uint // 0 表示未创建
	AdminRoleID       uint
	MenuCount         int
	ResourceCount     int
	DictCount         int
	DictItemCount     int
	ConfigGroupCount  int
	ConfigItemCount   int
	TenantUserSeqInit bool
	WarnMessages      []string // 非致命警告（如账号不存在）
}

// firstInstall 在 RunInPlatformTx 内执行：root org / admin role / 复制模板菜单资源字典 / admin user / tenant_user_seq。
//
// 全部走同一事务（bypass_rls=on），任一失败回滚，租户不留半成品。
//
// 复制来源：bootstrap 租户（migrations/framework.sql seed，admin 居住地）。
//   - menus 两步法复制（根 → 子，用 code 重映射 parent_id，再补 ancestors）
//   - resources 用 code 重映射 menu_id
//   - dicts + dict_items 用 code 重映射 dict_id
//   - admin role 绑定所有菜单 + 超级资源 *:*（resources 复制时已包含）
//
func firstInstall(ctx context.Context, pool *pgxpool.Pool, tenantID uint, adminAccountID uint) (*FirstInstallReport, error) {
	rep := &FirstInstallReport{}
	q, err := db.GetQuerier(ctx, pool)
	if err != nil {
		return nil, err
	}

	// 0) 查模板租户 ID
	var templateID uint
	err = q.QueryRow(ctx, `
		SELECT id FROM tenants
		WHERE code = $1 AND is_deleted = FALSE
		LIMIT 1`, TemplateTenantCode).Scan(&templateID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("模板租户 %q 不存在，请先执行 migrations/framework.sql", TemplateTenantCode)
		}
		return nil, fmt.Errorf("lookup template tenant: %w", err)
	}

	// 1) root 组织（每个租户必须有，作为后续组织的祖先）
	var rootOrgID uint
	if err := q.QueryRow(ctx, `
		INSERT INTO tenant_organizations (tenant_id, code, name, type, ancestors, sort, status)
		VALUES ($1, 'root', '根组织', 'company', '', 0, 1)
		RETURNING id`, tenantID).Scan(&rootOrgID); err != nil {
		return nil, fmt.Errorf("first-install root org: %w", err)
	}
	rep.RootOrgID = rootOrgID

	// 2) 复制 menus（两步法）
	// 2a) 根菜单（parent_id=0）：先入主，拿到新的 code → id 映射
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
		SELECT $1, code, name, subtitle, url, path, icon, sort, 0, '', visible, enabled
		FROM tenant_menus
		WHERE tenant_id = $2 AND parent_id = 0 AND is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy menus (roots): %w", err)
	}

	// 2b) 子菜单：用 code 重映射 parent_id 到新租户的同 code 菜单
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
		SELECT $1, m.code, m.name, m.subtitle, m.url, m.path, m.icon, m.sort,
		       new_p.id, '', m.visible, m.enabled
		FROM tenant_menus m
		JOIN tenant_menus old_p ON old_p.id = m.parent_id AND old_p.tenant_id = $2 AND old_p.is_deleted = FALSE
		JOIN tenant_menus new_p ON new_p.code = old_p.code
		                AND new_p.tenant_id = $1 AND new_p.is_deleted = FALSE
		WHERE m.tenant_id = $2 AND m.parent_id > 0 AND m.is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy menus (children): %w", err)
	}

	// 2c) 重建 ancestors（用 parent_id::text 表达层级路径）
	if _, err := q.Exec(ctx, `
		UPDATE tenant_menus SET ancestors = parent_id::text
		WHERE tenant_id = $1 AND parent_id > 0`, tenantID); err != nil {
		return nil, fmt.Errorf("first-install rebuild ancestors: %w", err)
	}

	// 记录菜单数量（用于 audit）
	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_menus WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID).Scan(&rep.MenuCount); err != nil {
		return nil, fmt.Errorf("first-install count menus: %w", err)
	}

	// 3) 复制 resources：用 code 重映射 menu_id
	// resources 表无 RLS，可任意 tenant_id 复制
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_permissions (tenant_id, menu_id, code, name, action, description, sort, status)
		SELECT $1, new_m.id, r.code, r.name, r.action, r.description, r.sort, r.status
		FROM tenant_permissions r
		LEFT JOIN tenant_menus new_m ON new_m.code = (
		        SELECT code FROM tenant_menus WHERE id = r.menu_id AND tenant_id = $2 AND is_deleted = FALSE
		    ) AND new_m.tenant_id = $1 AND new_m.is_deleted = FALSE
		WHERE r.tenant_id = $2 AND r.is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy resources: %w", err)
	}

	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_permissions WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID).Scan(&rep.ResourceCount); err != nil {
		return nil, fmt.Errorf("first-install count resources: %w", err)
	}

	// 4) admin 角色（data_scope=1 = DataScopeAll，独立内置）
	var adminRoleID uint
	if err := q.QueryRow(ctx, `
		INSERT INTO tenant_roles (tenant_id, code, name, description, data_scope, sort, status)
		VALUES ($1, 'admin', '系统管理员', '首装自动创建的内置超级管理员', 1, 1, 1)
		RETURNING id`, tenantID).Scan(&adminRoleID); err != nil {
		return nil, fmt.Errorf("first-install admin role: %w", err)
	}
	rep.AdminRoleID = adminRoleID

	// 4a) admin role 绑定所有模板菜单（role_menus 全绑定，确保租户内 admin 可访问所有菜单）
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_role_menus (tenant_id, role_id, menu_id)
		SELECT $1, $2, id FROM tenant_menus
		WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID, adminRoleID); err != nil {
		return nil, fmt.Errorf("first-install role_menus: %w", err)
	}

	// 4b) admin role 绑定超级资源 *:*（resources 复制时已包含此条，仅做绑定）
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_role_resources (tenant_id, role_id, resource_id, effect)
		SELECT $1, $2, id, 1
		FROM tenant_permissions
		WHERE tenant_id = $1 AND code = '*' AND action = '*' AND is_deleted = FALSE
		LIMIT 1`,
		tenantID, adminRoleID); err != nil {
		rep.WarnMessages = append(rep.WarnMessages,
			fmt.Sprintf("admin role 绑定超级资源失败: %v", err))
	}

	// 5) 复制 dicts：用 code 重映射（dict_items 的 dict_id 引用）
	if _, err := q.Exec(ctx, `
		INSERT INTO dicts (tenant_id, code, name, sort, status, extend)
		SELECT $1, code, name, sort, status, extend
		FROM dicts
		WHERE tenant_id = $2 AND is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy dicts: %w", err)
	}

	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM dicts WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID).Scan(&rep.DictCount); err != nil {
		return nil, fmt.Errorf("first-install count dicts: %w", err)
	}

	// 6) 复制 dict_items：用 code 重映射 dict_id
	if _, err := q.Exec(ctx, `
		INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend)
		SELECT $1, new_d.id, di.code, di.name, di.sort, di.status, di.extend
		FROM dict_items di
		JOIN dicts old_d ON old_d.id = di.dict_id AND old_d.tenant_id = $2 AND old_d.is_deleted = FALSE
		JOIN dicts new_d ON new_d.code = old_d.code
		                AND new_d.tenant_id = $1 AND new_d.is_deleted = FALSE
		WHERE di.tenant_id = $2 AND di.is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy dict_items: %w", err)
	}

	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM dict_items WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID).Scan(&rep.DictItemCount); err != nil {
		return nil, fmt.Errorf("first-install count dict_items: %w", err)
	}

	// 6a) 复制 config_categories（先入主）
	if _, err := q.Exec(ctx, `
		INSERT INTO config_categories (tenant_id, code, name, description, icon, sort, is_system, is_public, status)
		SELECT $1, code, name, description, icon, sort, is_system, is_public, status
		FROM config_categories
		WHERE tenant_id = $2 AND is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy config_categories: %w", err)
	}

	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM config_categories WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID).Scan(&rep.ConfigGroupCount); err != nil {
		return nil, fmt.Errorf("first-install count config_categories: %w", err)
	}

	// 6b) 复制 config_items（用 code 重映射 category_id，value 继承 default_value）
	if _, err := q.Exec(ctx, `
		INSERT INTO config_items
		    (tenant_id, category_id, key, value, default_value, type, label, description, options, validation,
		     sort, is_public, is_readonly, is_system, status)
		SELECT $1, new_g.id, ci.key, COALESCE(ci.default_value, ci.value), ci.default_value, ci.type,
		       ci.label, ci.description, ci.options, ci.validation,
		       ci.sort, ci.is_public, ci.is_readonly, ci.is_system, ci.status
		FROM config_items ci
		JOIN config_categories old_g ON old_g.id = ci.category_id AND old_g.tenant_id = $2 AND old_g.is_deleted = FALSE
		JOIN config_categories new_g ON new_g.code = old_g.code
		                        AND new_g.tenant_id = $1 AND new_g.is_deleted = FALSE
		WHERE ci.tenant_id = $2 AND ci.is_deleted = FALSE`,
		tenantID, templateID); err != nil {
		return nil, fmt.Errorf("first-install copy config_items: %w", err)
	}

	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM config_items WHERE tenant_id = $1 AND is_deleted = FALSE`,
		tenantID).Scan(&rep.ConfigItemCount); err != nil {
		return nil, fmt.Errorf("first-install count config_items: %w", err)
	}

	// 7) admin user（仅当 adminAccountID > 0）
	if adminAccountID > 0 {
		if err := installAdminUser(ctx, q, tenantID, adminAccountID, rootOrgID, adminRoleID, rep); err != nil {
			return nil, err
		}
	}

	// 8) tenant_user_seq 初始化（用于 user_code 自增）
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_user_seq (tenant_id, last_seq)
		VALUES ($1, 0)
		ON CONFLICT (tenant_id) DO NOTHING`, tenantID); err != nil {
		if !isUniqueViolation(err) {
			return nil, fmt.Errorf("first-install tenant_user_seq: %w", err)
		}
	}
	rep.TenantUserSeqInit = true

	return rep, nil
}

// installAdminUser 创建 admin 用户并绑定到 admin role。
// 单独抽出便于处理"账号缺失仅 warn"的非致命分支。
func installAdminUser(ctx context.Context, q db.Querier, tenantID, adminAccountID, rootOrgID, adminRoleID uint, rep *FirstInstallReport) error {
	// 7a) 校验账号存在且 status=1
	var accountStatus int16
	err := q.QueryRow(ctx, `
		SELECT status FROM accounts
		WHERE id = $1 AND is_deleted = FALSE`, adminAccountID).Scan(&accountStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			rep.WarnMessages = append(rep.WarnMessages,
				fmt.Sprintf("admin_account_id=%d 不存在，跳过 admin user 创建", adminAccountID))
			return nil
		}
		return fmt.Errorf("first-install check admin account: %w", err)
	}
	if accountStatus != 1 {
		rep.WarnMessages = append(rep.WarnMessages,
			fmt.Sprintf("admin_account_id=%d status=%d，跳过 admin user 创建", adminAccountID, accountStatus))
		return nil
	}

	// 7b) INSERT admin user（uk_users_account 是 partial unique index on account_id WHERE is_deleted=FALSE）
	// ON CONFLICT 兜底重查；同一账号可能被多租户共享。
	var adminUserID uint
	err = q.QueryRow(ctx, `
		INSERT INTO tenant_users (tenant_id, account_id, org_id, real_name, status)
		VALUES ($1, $2, $3, 'Administrator', 1)
		ON CONFLICT (account_id) WHERE is_deleted = FALSE
		DO UPDATE SET org_id = EXCLUDED.org_id, updated_at = NOW()
		RETURNING id`, tenantID, adminAccountID, rootOrgID).Scan(&adminUserID)
	if err != nil {
		if isUniqueViolation(err) {
			if err := q.QueryRow(ctx, `
				SELECT id FROM tenant_users WHERE account_id = $1 AND is_deleted = FALSE LIMIT 1`,
				adminAccountID).Scan(&adminUserID); err != nil {
				return fmt.Errorf("first-install locate existing user: %w", err)
			}
		} else {
			return fmt.Errorf("first-install admin user: %w", err)
		}
	}
	rep.AdminUserID = adminUserID

	// 7c) admin user ↔ admin role 绑定
	if _, err := q.Exec(ctx, `
		INSERT INTO tenant_user_roles (tenant_id, user_id, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) WHERE is_deleted = FALSE DO NOTHING`,
		tenantID, adminUserID, adminRoleID); err != nil {
		return fmt.Errorf("first-install user_roles: %w", err)
	}
	return nil
}

// isUniqueViolation PG 唯一约束冲突判断。23505 = unique_violation。
// 用于 ON CONFLICT 兜底分支——partial index 上 DO UPDATE 可能不触发，需手动判断重查。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

package permission

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// PostgresPermissionRepository implements PermissionRepository
type PostgresPermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{db: db}
}

// allActions 列出了 resource 通配符 * 展开时注入的 action。
// 与 constants.go 的 Act* 常量保持同步；中间件 HasPermission 查的是同样的 key。
var allActions = []string{ActList, ActGet, ActCreate, ActUpdate, ActDelete, ActTree}

// expandPermissionCode 把单条 permission code 展开成一组运行时 map key。
// 规则（与 apps/sys/permission/service.go permissionCodeValid 对齐）：
//   - "x"    → {"x", "x:*"}                                                菜单无关资源（无 action 维度）
//   - "x:y"  → {"x:y"}                                                      菜单相关资源（具体 action）
//   - "x:*"  → {"x:list","x:get","x:create","x:update","x:delete","x:tree"} 菜单相关资源（所有 action）
//   - "*:*"  → {"*:*"}                                                      全局通配（admin 专用）
//   - 其他格式（多冒号、空段）应被 service 层拦截，到不了这里。
//
// 展开 "x" 为 "x" 和 "x:*" 是为了中间件 HasPermission("x", any_action) 走通配路径匹配。
func expandPermissionCode(code string) []string {
	if code == "*:*" {
		return []string{"*:*"}
	}
	if strings.HasSuffix(code, ":*") {
		resource := strings.TrimSuffix(code, ":*")
		keys := make([]string, 0, len(allActions))
		for _, act := range allActions {
			keys = append(keys, resource+":"+act)
		}
		return keys
	}
	if !strings.Contains(code, ":") {
		// 纯字符串：菜单无关资源。x:* 让中间件 HasPermission 走通配路径匹配所有 action。
		return []string{code, code + ":*"}
	}
	return []string{code}
}

// GetUserPermissions returns map of "resource:action" -> true.
//
// 路径合集：tenant 域 + sys 域。
//
//	(1) tenant 域：userID = tenant_users.id
//	    tenant_users -> tenant_user_roles -> tenant_roles
//	    -> tenant_role_resources -> tenant_permissions
//	(2) sys 域：userID = sys_users.account_id（per JWT 注释：sys admin 的
//	    Claims.UserID = account_id；同一个人可能同时拥有 tenant 身份与 sys 身份，
//	    但 userID 解析到的表不同，所以 UNOIN ALL 不会重复）
//
// 约定（0024+ 终态）：
//   - tenant_permissions.code / sys_permissions.code 均为完整串 "resource:action"，
//     存储层禁止两段式（service 层 permissionCodeValid 拦截）。
//   - 运行时 map key 直接用 code；不拼装、不 split_part。
//   - "x:*" 表示该资源所有操作（allActions 展开）；"*:*" 表示全局通配（admin 专用）。
//
// 变更记录：
//   - 0024：两域统一约定（不再 split_part、不再 code+":"+action 拼装）。
//     统一 SQL 只选 code 列；客户端用 expandPermissionCode 处理通配展开。
func (r *PostgresPermissionRepository) GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		-- (1) tenant 域权限
		SELECT DISTINCT res.code
		FROM tenant_users u
		JOIN tenant_user_roles ur ON ur.user_id = u.id
		JOIN tenant_roles rol ON rol.id = ur.role_id
		JOIN tenant_role_resources rr ON rr.role_id = rol.id
		JOIN tenant_permissions res ON res.id = rr.permission_id
		WHERE u.id = $1
		  AND u.is_deleted = FALSE
		  AND ur.is_deleted = FALSE
		  AND rol.is_deleted = FALSE
		  AND rol.status = 1
		  AND rr.is_deleted = FALSE
		  AND rr.effect = 1
		  AND res.is_deleted = FALSE
		  AND res.status = 1

		UNION ALL

				-- (2) sys 域权限
		SELECT DISTINCT p.code
		FROM sys_users su
		JOIN sys_user_roles sur ON sur.user_id = su.id
		JOIN sys_roles r ON r.id = sur.role_id
		JOIN sys_role_permissions rp ON rp.role_id = r.id
		JOIN sys_permissions p ON p.id = rp.permission_id
		WHERE su.account_id = $1
		  AND su.is_deleted = FALSE
		  AND sur.is_deleted = FALSE
		  AND r.is_deleted = FALSE
		  AND r.status = 1
		  AND rp.is_deleted = FALSE
		  AND rp.effect = 1
		  AND p.is_deleted = FALSE
		  AND p.status = 1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	perms := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		for _, key := range expandPermissionCode(code) {
			perms[key] = true
		}
	}
	return perms, nil
}

// GetUserRoles returns role codes for a user
func (r *PostgresPermissionRepository) GetUserRoles(ctx context.Context, userID uint) ([]string, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT DISTINCT rol.code
		FROM tenant_user_roles ur
		JOIN tenant_roles rol ON rol.id = ur.role_id
		WHERE ur.user_id = $1
		  AND ur.is_deleted = FALSE
		  AND rol.is_deleted = FALSE
		  AND rol.status = 1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		roles = append(roles, code)
	}
	return roles, nil
}

// GetUserIDsByRole returns all user IDs that have the given role
func (r *PostgresPermissionRepository) GetUserIDsByRole(ctx context.Context, roleID uint) ([]uint, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT ur.user_id
		FROM tenant_user_roles ur
		WHERE ur.role_id = $1
		  AND ur.is_deleted = FALSE
	`, roleID)
	if err != nil {
		return nil, fmt.Errorf("get user ids by role: %w", err)
	}
	defer rows.Close()

	var userIDs []uint
	for rows.Next() {
		var uid uint
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, uid)
	}
	return userIDs, nil
}

// GetUserIDsByResource returns all user IDs whose roles reference the given resource
func (r *PostgresPermissionRepository) GetUserIDsByResource(ctx context.Context, resourceID uint) ([]uint, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT DISTINCT ur.user_id
		FROM tenant_role_resources rr
		JOIN tenant_user_roles ur ON ur.role_id = rr.role_id AND ur.is_deleted = FALSE
		JOIN tenant_roles rol ON rol.id = rr.role_id AND rol.is_deleted = FALSE AND rol.status = 1
		WHERE rr.permission_id = $1
		  AND rr.is_deleted = FALSE
		  AND rr.effect = 1
	`, resourceID)
	if err != nil {
		return nil, fmt.Errorf("get user ids by resource: %w", err)
	}
	defer rows.Close()

	var userIDs []uint
	for rows.Next() {
		var uid uint
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, uid)
	}
	return userIDs, nil
}

// PostgresDataScopeRepository implements DataScopeRepository
type PostgresDataScopeRepository struct {
	db *pgxpool.Pool
}

func NewDataScopeRepository(db *pgxpool.Pool) *PostgresDataScopeRepository {
	return &PostgresDataScopeRepository{db: db}
}

// GetDataScope returns the data scope for a user based on their roles
// Takes the most permissive scope if user has multiple roles
func (r *PostgresDataScopeRepository) GetDataScope(ctx context.Context, userID uint) (*DataScope, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	// Get the most permissive data_scope from user's roles
	// data_scope: 1=全部 > 4=本部门及以下 > 3=本部门 > 2=自定义 > 5=本人
	var dataScope int
	err = q.QueryRow(ctx, `
		SELECT COALESCE(MIN(rol.data_scope), 5)
		FROM tenant_user_roles ur
		JOIN tenant_roles rol ON rol.id = ur.role_id
		WHERE ur.user_id = $1
		  AND ur.is_deleted = FALSE
		  AND rol.is_deleted = FALSE
		  AND rol.status = 1
	`, userID).Scan(&dataScope)
	if err != nil {
		return nil, fmt.Errorf("get user data scope: %w", err)
	}

	ds := &DataScope{
		Type: DataScopeType(dataScope),
	}

	// For custom data scope (type=2), load the allowed org_ids
	if ds.Type == DataScopeCustom {
		rows, err := q.Query(ctx, `
			SELECT rds.org_id
			FROM tenant_user_roles ur
			JOIN tenant_role_data_scopes rds ON rds.role_id = ur.role_id
			WHERE ur.user_id = $1
			  AND ur.is_deleted = FALSE
		`, userID)
		if err != nil {
			return nil, fmt.Errorf("get custom org ids: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var oid int64
			if err := rows.Scan(&oid); err != nil {
				return nil, err
			}
			ds.OrgIDs = append(ds.OrgIDs, oid)
		}
	}

	return ds, nil
}

// GetUserOrgID returns the user's organization ID
func (r *PostgresDataScopeRepository) GetUserOrgID(ctx context.Context, userID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var orgID int64
	err = q.QueryRow(ctx, `
		SELECT org_id FROM tenant_users WHERE id = $1 AND is_deleted = FALSE
	`, userID).Scan(&orgID)
	if err != nil {
		// org_id can be NULL for some users
		return 0, nil
	}
	return orgID, nil
}

// GetByRoleID returns org_ids for a role's custom data scope
func (r *PostgresDataScopeRepository) GetByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT org_id FROM tenant_role_data_scopes WHERE role_id = $1
	`, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role data scopes: %w", err)
	}
	defer rows.Close()

	var orgIDs []uint
	for rows.Next() {
		var oid int64
		if err := rows.Scan(&oid); err != nil {
			return nil, err
		}
		orgIDs = append(orgIDs, uint(oid))
	}
	return orgIDs, nil
}

// SetForRole replaces all data scopes for a role
func (r *PostgresDataScopeRepository) SetForRole(ctx context.Context, roleID uint, orgIDs []uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	// Delete existing
	_, err = q.Exec(ctx, `DELETE FROM tenant_role_data_scopes WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("delete existing: %w", err)
	}

	// Get tenant_id for the role
	var tenantID int64
	err = q.QueryRow(ctx, `SELECT tenant_id FROM tenant_roles WHERE id = $1`, roleID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("get tenant_id: %w", err)
	}

	// Insert new (batch)
	if len(orgIDs) > 0 {
		_, err = q.Exec(ctx, `
			INSERT INTO tenant_role_data_scopes (tenant_id, role_id, org_id)
			SELECT $1, $2, unnest
			FROM unnest($3::bigint[]) AS unnest
		`, tenantID, roleID, orgIDs)
		if err != nil {
			return fmt.Errorf("insert data scopes: %w", err)
		}
	}

	return nil
}

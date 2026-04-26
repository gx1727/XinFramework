package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/permission"
)

// PostgresPermissionRepository implements permission.PermissionRepository
type PostgresPermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{db: db}
}

// GetUserPermissions returns map of "resource:action" -> true
// Joins through: users -> user_roles -> roles -> permissions -> resources
func (r *PostgresPermissionRepository) GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT res.code, res.action
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles rol ON rol.id = ur.role_id
		JOIN permissions perm ON perm.role_id = rol.id
		JOIN resources res ON res.id = perm.resource_id
		WHERE u.id = $1
		  AND u.is_deleted = FALSE
		  AND ur.is_deleted = FALSE
		  AND rol.is_deleted = FALSE
		  AND rol.status = 1
		  AND perm.is_deleted = FALSE
		  AND perm.effect = 1
		  AND res.is_deleted = FALSE
		  AND res.status = 1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	perms := make(map[string]bool)
	for rows.Next() {
		var code, action string
		if err := rows.Scan(&code, &action); err != nil {
			return nil, err
		}
		// Format: "resource_code:action" e.g., "user:create"
		key := code + ":" + action
		perms[key] = true
	}
	return perms, nil
}

// GetUserRoles returns role codes for a user
func (r *PostgresPermissionRepository) GetUserRoles(ctx context.Context, userID uint) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT rol.code
		FROM user_roles ur
		JOIN roles rol ON rol.id = ur.role_id
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

// PostgresDataScopeRepository implements permission.DataScopeRepository
type PostgresDataScopeRepository struct {
	db *pgxpool.Pool
}

func NewDataScopeRepository(db *pgxpool.Pool) *PostgresDataScopeRepository {
	return &PostgresDataScopeRepository{db: db}
}

// GetDataScope returns the data scope for a user based on their roles
// Takes the most permissive scope if user has multiple roles
func (r *PostgresDataScopeRepository) GetDataScope(ctx context.Context, userID uint) (*permission.DataScope, error) {
	// Get the most permissive data_scope from user's roles
	// data_scope: 1=全部 > 4=本部门及以下 > 3=本部门 > 2=自定义 > 5=本人
	var dataScope int
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(MIN(rol.data_scope), 5)
		FROM user_roles ur
		JOIN roles rol ON rol.id = ur.role_id
		WHERE ur.user_id = $1
		  AND ur.is_deleted = FALSE
		  AND rol.is_deleted = FALSE
		  AND rol.status = 1
	`, userID).Scan(&dataScope)
	if err != nil {
		return nil, fmt.Errorf("get user data scope: %w", err)
	}

	ds := &permission.DataScope{
		Type: permission.DataScopeType(dataScope),
	}

	// For custom data scope (type=2), load the allowed org_ids
	if ds.Type == permission.DataScopeCustom {
		rows, err := r.db.Query(ctx, `
			SELECT rds.org_id
			FROM user_roles ur
			JOIN role_data_scopes rds ON rds.role_id = ur.role_id
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
	var orgID int64
	err := r.db.QueryRow(ctx, `
		SELECT org_id FROM users WHERE id = $1 AND is_deleted = FALSE
	`, userID).Scan(&orgID)
	if err != nil {
		// org_id can be NULL for some users
		return 0, nil
	}
	return orgID, nil
}

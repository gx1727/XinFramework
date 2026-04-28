package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/permission"
)

type PostgresRolePermissionRepository struct {
	db *pgxpool.Pool
}

func NewRolePermissionRepository(db *pgxpool.Pool) permission.PermissionRepository {
	return &PostgresRolePermissionRepository{db: db}
}

func (r *PostgresRolePermissionRepository) GetByRoleID(ctx context.Context, roleID uint) ([]permission.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, role_id, resource_type, resource_id, resource_code, effect
		FROM permissions
		WHERE is_deleted = FALSE AND role_id = $1
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []permission.Permission
	for rows.Next() {
		var p permission.Permission
		if err := rows.Scan(&p.ID, &p.TenantID, &p.RoleID, &p.ResourceType, &p.ResourceID, &p.ResourceCode, &p.Effect); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *PostgresRolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	_, err := r.db.Exec(ctx, `UPDATE permissions SET is_deleted = TRUE, updated_at = NOW() WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("delete permissions: %w", err)
	}
	return nil
}

func (r *PostgresRolePermissionRepository) Create(ctx context.Context, tenantID, roleID uint, p permission.Permission) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO permissions (tenant_id, role_id, resource_type, resource_id, resource_code, effect)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tenantID, roleID, p.ResourceType, p.ResourceID, p.ResourceCode, p.Effect)
	if err != nil {
		return fmt.Errorf("create permission: %w", err)
	}
	return nil
}

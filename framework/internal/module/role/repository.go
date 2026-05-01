package role

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRoleRepository implements RoleRepository

type PostgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uint) (*Role, error) {
	var role Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
		FROM roles
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*Role, error) {
	var role Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
		FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.tenant_id, r.org_id, r.code, r.name, r.description, r.data_scope, r.extend, r.is_default, r.sort, r.status, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.is_deleted = FALSE`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var extend []byte
		if err := rows.Scan(
			&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
			&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if extend != nil {
			role.Extend = string(extend)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Role, int64, error) {
	offset := (page - 1) * size
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
		FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1
		AND ($2 = '' OR (code ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%'))
		ORDER BY sort ASC, id ASC
		LIMIT $3 OFFSET $4`, tenantID, keyword, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var extend []byte
		if err := rows.Scan(
			&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
			&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if extend != nil {
			role.Extend = string(extend)
		}
		roles = append(roles, role)
	}

	var total int64
	r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1
		AND ($2 = '' OR (code ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%'))`,
		tenantID, keyword).Scan(&total)

	return roles, total, nil
}

func (r *PostgresRoleRepository) Create(ctx context.Context, tenantID uint, req CreateRoleRepoReq) (*Role, error) {
	var role Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		INSERT INTO roles (tenant_id, code, name, description, data_scope, is_default, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
	`, tenantID, req.Code, req.Name, req.Description, req.DataScope, req.IsDefault, req.Sort, req.Status).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) Update(ctx context.Context, id uint, req UpdateRoleRepoReq) (*Role, error) {
	var role Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		UPDATE roles SET name = $1, description = $2, data_scope = $3, is_default = $4, sort = $5, status = $6, updated_at = NOW()
		WHERE id = $7 AND is_deleted = FALSE
		RETURNING id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
	`, req.Name, req.Description, req.DataScope, req.IsDefault, req.Sort, req.Status, id).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("update role: %w", err)
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id uint) error {
	_, err := r.db.Exec(ctx, `UPDATE roles SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

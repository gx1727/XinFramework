package role

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// PostgresRoleRepository implements RoleRepository

type PostgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uint) (*Role, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var role Role
	var extend []byte
	err = q.QueryRow(ctx, `
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var role Role
	var extend []byte
	err = q.QueryRow(ctx, `
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	rows, err := q.Query(ctx, `
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
	q.QueryRow(ctx, `
		SELECT COUNT(*) FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1
		AND ($2 = '' OR (code ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%'))`,
		tenantID, keyword).Scan(&total)

	return roles, total, nil
}

func (r *PostgresRoleRepository) Create(ctx context.Context, tenantID uint, req CreateRoleRepoReq) (*Role, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var role Role
	var extend []byte
	err = q.QueryRow(ctx, `
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var role Role
	var extend []byte
	err = q.QueryRow(ctx, `
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

// Patch 局部更新：仅修改 req 中非 nil 的字段，nil 字段保持原值
func (r *PostgresRoleRepository) Patch(ctx context.Context, id uint, req PatchRoleRepoReq) (*Role, error) {
	sets := make([]string, 0, 6)
	args := make([]interface{}, 0, 7)
	idx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *req.Name)
		idx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", idx))
		args = append(args, *req.Description)
		idx++
	}
	if req.DataScope != nil {
		sets = append(sets, fmt.Sprintf("data_scope = $%d", idx))
		args = append(args, *req.DataScope)
		idx++
	}
	if req.IsDefault != nil {
		sets = append(sets, fmt.Sprintf("is_default = $%d", idx))
		args = append(args, *req.IsDefault)
		idx++
	}
	if req.Sort != nil {
		sets = append(sets, fmt.Sprintf("sort = $%d", idx))
		args = append(args, *req.Sort)
		idx++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", idx))
		args = append(args, *req.Status)
		idx++
	}

	// 未提供任何字段 → 直接返回当前记录
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`
		UPDATE roles SET %s
		WHERE id = $%d AND is_deleted = FALSE
		RETURNING id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
	`, strings.Join(sets, ", "), idx)

	var role Role
	var extend []byte
	if err := q.QueryRow(ctx, sql, args...).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("patch role: %w", err)
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE roles SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

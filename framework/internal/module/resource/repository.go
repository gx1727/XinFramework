package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresResourceRepository struct {
	db *pgxpool.Pool
}

func NewResourceRepository(db *pgxpool.Pool) ResourceRepository {
	return &PostgresResourceRepository{db: db}
}

func (r *PostgresResourceRepository) GetByID(ctx context.Context, id uint) (*Resource, error) {
	var res Resource
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, menu_id, code, name, action, description, sort, status, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *PostgresResourceRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*Resource, error) {
	var res Resource
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, menu_id, code, name, action, description, sort, status, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *PostgresResourceRepository) GetByTenant(ctx context.Context, tenantID uint) ([]Resource, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, menu_id, code, name, action, description, sort, status, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (r *PostgresResourceRepository) GetByMenu(ctx context.Context, menuID uint) ([]Resource, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, menu_id, code, name, action, description, sort, status, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND menu_id = $1
		ORDER BY sort ASC, id ASC`, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (r *PostgresResourceRepository) GetUserResources(ctx context.Context, tenantID, userID uint) ([]Resource, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT r.id, r.tenant_id, r.menu_id, r.code, r.name, r.action, r.description, r.sort, r.status, r.created_at, r.updated_at
		FROM resources r
		JOIN permissions p ON p.resource_type = 'resource' AND p.resource_code = r.code
		JOIN user_roles ur ON ur.role_id = p.role_id
		WHERE r.is_deleted = FALSE AND r.tenant_id = $1 AND ur.user_id = $2 AND ur.is_deleted = FALSE
		ORDER BY r.id ASC`, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var res Resource
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (r *PostgresResourceRepository) Create(ctx context.Context, tenantID uint, req CreateResourceRepoReq) (*Resource, error) {
	var res Resource
	err := r.db.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, menu_id, code, name, action, description, sort, status, created_at, updated_at
	`, tenantID, req.MenuID, req.Code, req.Name, req.Action, req.Description, req.Sort, req.Status).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uk_resource_code") {
			return nil, errors.New("resource code already exists")
		}
		return nil, fmt.Errorf("create resource: %w", err)
	}
	return &res, nil
}

func (r *PostgresResourceRepository) Update(ctx context.Context, id uint, req UpdateResourceRepoReq) (*Resource, error) {
	var res Resource
	err := r.db.QueryRow(ctx, `
		UPDATE resources SET name = $2, action = $3, description = $4, sort = $5, status = $6, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, menu_id, code, name, action, description, sort, status, created_at, updated_at
	`, id, req.Name, req.Action, req.Description, req.Sort, req.Status).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Action, &res.Description, &res.Sort, &res.Status,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("update resource: %w", err)
	}
	return &res, nil
}

func (r *PostgresResourceRepository) Delete(ctx context.Context, id uint) error {
	tag, err := r.db.Exec(ctx, `UPDATE resources SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete resource: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

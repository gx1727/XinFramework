package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/model"
)

type PostgresMenuRepository struct {
	db *pgxpool.Pool
}

func NewMenuRepository(db *pgxpool.Pool) model.MenuRepository {
	return &PostgresMenuRepository{db: db}
}

func (r *PostgresMenuRepository) GetByID(ctx context.Context, id uint) (*model.Menu, error) {
	var m model.Menu
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, created_at, updated_at
		FROM menus
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrMenuNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*model.Menu, error) {
	var m model.Menu
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, created_at, updated_at
		FROM menus
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrMenuNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetByTenant(ctx context.Context, tenantID uint) ([]model.Menu, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, created_at, updated_at
		FROM menus
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []model.Menu
	for rows.Next() {
		var m model.Menu
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
			&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetUserMenus(ctx context.Context, tenantID, userID uint) ([]model.Menu, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT m.id, m.tenant_id, m.code, m.name, m.subtitle, m.url, m.path, m.icon, m.sort, m.parent_id, m.ancestors, m.created_at, m.updated_at
		FROM menus m
		JOIN permissions p ON p.resource_type = 'menu' AND p.resource_code = m.code
		JOIN user_roles ur ON ur.role_id = p.role_id
		WHERE m.is_deleted = FALSE AND m.tenant_id = $1 AND ur.user_id = $2 AND ur.is_deleted = FALSE
		ORDER BY m.sort ASC, m.id ASC`, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []model.Menu
	for rows.Next() {
		var m model.Menu
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
			&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}

type PostgresResourceRepository struct {
	db *pgxpool.Pool
}

func NewResourceRepository(db *pgxpool.Pool) model.ResourceRepository {
	return &PostgresResourceRepository{db: db}
}

func (r *PostgresResourceRepository) GetByID(ctx context.Context, id uint) (*model.Resource, error) {
	var res model.Resource
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, menu_id, code, name, description, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Description,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrResourceNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *PostgresResourceRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*model.Resource, error) {
	var res model.Resource
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, menu_id, code, name, description, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Description,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrResourceNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *PostgresResourceRepository) GetByTenant(ctx context.Context, tenantID uint) ([]model.Resource, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, menu_id, code, name, description, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []model.Resource
	for rows.Next() {
		var res model.Resource
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Description,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (r *PostgresResourceRepository) GetUserResources(ctx context.Context, tenantID, userID uint) ([]model.Resource, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT r.id, r.tenant_id, r.menu_id, r.code, r.name, r.description, r.created_at, r.updated_at
		FROM resources r
		JOIN permissions p ON p.resource_type = 'resource' AND p.resource_code = r.code
		JOIN user_roles ur ON ur.role_id = p.role_id
		WHERE r.is_deleted = FALSE AND r.tenant_id = $1 AND ur.user_id = $2 AND ur.is_deleted = FALSE
		ORDER BY r.id ASC`, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []model.Resource
	for rows.Next() {
		var res model.Resource
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Description,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

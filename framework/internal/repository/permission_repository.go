package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at
		FROM menus
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
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
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at
		FROM menus
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
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
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at
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
			&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) GetUserMenus(ctx context.Context, tenantID, userID uint) ([]model.Menu, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT m.id, m.tenant_id, m.code, m.name, m.subtitle, m.url, m.path, m.icon, m.sort, m.parent_id, m.ancestors, m.visible, m.enabled, m.created_at, m.updated_at
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
			&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}

func (r *PostgresMenuRepository) Create(ctx context.Context, tenantID uint, req model.CreateMenuRepoReq) (*model.Menu, error) {
	var m model.Menu
	err := r.db.QueryRow(ctx, `
		INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at`,
		tenantID, req.Code, req.Name, req.Subtitle, req.URL, req.Path, req.Icon, req.Sort, req.ParentID, req.Ancestors, req.Visible, req.Enabled,
	).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uk_menu_code") {
			return nil, errors.New("menu code already exists")
		}
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return &m, nil
}

func (r *PostgresMenuRepository) Update(ctx context.Context, id uint, req model.UpdateMenuRepoReq) (*model.Menu, error) {
	var m model.Menu
	err := r.db.QueryRow(ctx, `
		UPDATE menus SET
			code = $2, name = $3, subtitle = $4, url = $5, path = $6, icon = $7, sort = $8, visible = $9, enabled = $10, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at`,
		id, req.Code, req.Name, req.Subtitle, req.URL, req.Path, req.Icon, req.Sort, req.Visible, req.Enabled,
	).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrMenuNotFound
		}
		if strings.Contains(err.Error(), "uk_menu_code") {
			return nil, errors.New("menu code already exists")
		}
		return nil, fmt.Errorf("update menu: %w", err)
	}
	return &m, nil
}

func (r *PostgresMenuRepository) Delete(ctx context.Context, id uint) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE menus SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrMenuNotFound
	}
	return nil
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

func (r *PostgresResourceRepository) GetByMenu(ctx context.Context, menuID uint) ([]model.Resource, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, menu_id, code, name, description, created_at, updated_at
		FROM resources
		WHERE is_deleted = FALSE AND menu_id = $1
		ORDER BY sort ASC, id ASC`, menuID)
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

func (r *PostgresResourceRepository) Create(ctx context.Context, tenantID uint, req model.CreateResourceRepoReq) (*model.Resource, error) {
	var res model.Resource
	err := r.db.QueryRow(ctx, `
		INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, menu_id, code, name, description, created_at, updated_at
	`, tenantID, req.MenuID, req.Code, req.Name, req.Action, req.Description, req.Sort, req.Status).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Description,
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

func (r *PostgresResourceRepository) Update(ctx context.Context, id uint, req model.UpdateResourceRepoReq) (*model.Resource, error) {
	var res model.Resource
	err := r.db.QueryRow(ctx, `
		UPDATE resources SET name = $2, action = $3, description = $4, sort = $5, status = $6, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, menu_id, code, name, description, created_at, updated_at
	`, id, req.Name, req.Action, req.Description, req.Sort, req.Status).Scan(
		&res.ID, &res.TenantID, &res.MenuID, &res.Code, &res.Name, &res.Description,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrResourceNotFound
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
		return model.ErrResourceNotFound
	}
	return nil
}

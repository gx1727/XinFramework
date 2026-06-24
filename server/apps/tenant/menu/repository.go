package menu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

type PostgresMenuRepository struct {
	db *pgxpool.Pool
}

func NewMenuRepository(db *pgxpool.Pool) MenuRepository {
	return &PostgresMenuRepository{db: db}
}

func (r *PostgresMenuRepository) GetByID(ctx context.Context, id uint) (_ *Menu, err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var m Menu
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at
		FROM tenant_menus
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetByCode(ctx context.Context, tenantID uint, code string) (_ *Menu, err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var m Menu
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at
		FROM tenant_menus
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetByTenant(ctx context.Context, tenantID uint) (_ []Menu, err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at
		FROM tenant_menus
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
			&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
			&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	// 检查遍历过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return menus, nil
}

func (r *PostgresMenuRepository) GetUserMenus(ctx context.Context, tenantID, userID uint) (_ []Menu, err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT DISTINCT m.id, m.tenant_id, m.code, m.name, m.subtitle, m.url, m.path, m.icon, m.sort, m.parent_id, m.ancestors, m.visible, m.enabled, m.created_at, m.updated_at
		FROM tenant_menus m
		JOIN tenant_role_menus rm ON rm.menu_id = m.id AND rm.is_deleted = FALSE
		JOIN tenant_user_roles ur ON ur.role_id = rm.role_id AND ur.is_deleted = FALSE
		WHERE m.is_deleted = FALSE AND m.tenant_id = $1 AND ur.user_id = $2
		ORDER BY m.sort ASC, m.id ASC`, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
			&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
			&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	// 检查遍历过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return menus, nil
}

func (r *PostgresMenuRepository) Create(ctx context.Context, tenantID uint, req CreateMenuRepoReq) (_ *Menu, err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var m Menu
	err = q.QueryRow(ctx, `
		INSERT INTO tenant_menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
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
			return nil, ErrMenuCodeExistsDB
		}
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return &m, nil
}

func (r *PostgresMenuRepository) Update(ctx context.Context, id uint, req UpdateMenuRepoReq) (_ *Menu, err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var m Menu
	err = q.QueryRow(ctx, `
		UPDATE tenant_menus SET
			code = $2, name = $3, subtitle = $4, url = $5, path = $6, icon = $7, sort = $8, parent_id = $9, ancestors = $10, visible = $11, enabled = $12, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at`,
		id, req.Code, req.Name, req.Subtitle, req.URL, req.Path, req.Icon, req.Sort, req.ParentID, req.Ancestors, req.Visible, req.Enabled,
	).Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMenuNotFound
		}
		if strings.Contains(err.Error(), "uk_menu_code") {
			return nil, ErrMenuCodeExistsDB
		}
		return nil, fmt.Errorf("update menu: %w", err)
	}
	return &m, nil
}

func (r *PostgresMenuRepository) Delete(ctx context.Context, tenantID, id uint) (err error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, "SELECT set_config('app.show_deleted', $1, true)", "true")
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE tenant_menus SET is_deleted = TRUE, updated_at = NOW(), tenant_id = $2
		WHERE is_deleted = FALSE AND id = $1`, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMenuNotFound
	}
	return nil
}

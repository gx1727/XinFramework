package platformmenu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

// platformTenantID 是平台菜单的"租户"标识。
//
// 所有 SQL 必须用这个常量 —— 严禁从调用方传入 tenant_id。
// 这是 platform_menu 强制性的不变量。
const platformTenantID uint = 0

type PostgresMenuRepository struct {
	db *pgxpool.Pool
}

func NewMenuRepository(pool *pgxpool.Pool) MenuRepository {
	return &PostgresMenuRepository{db: pool}
}

const menuColumns = `id, tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at`

func scanMenu(row pgx.Row, m *Menu) error {
	return row.Scan(
		&m.ID, &m.TenantID, &m.Code, &m.Name, &m.Subtitle,
		&m.URL, &m.Path, &m.Icon, &m.Sort, &m.ParentID, &m.Ancestors,
		&m.Visible, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
}

func (r *PostgresMenuRepository) GetByID(ctx context.Context, id uint) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m Menu
	err = scanMenu(q.QueryRow(ctx, `
		SELECT `+menuColumns+`
		FROM menus
		WHERE is_deleted = FALSE AND tenant_id = $1 AND id = $2`,
		platformTenantID, id), &m)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errMenuNotFoundDB
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetByCode(ctx context.Context, code string) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m Menu
	err = scanMenu(q.QueryRow(ctx, `
		SELECT `+menuColumns+`
		FROM menus
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`,
		platformTenantID, code), &m)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errMenuNotFoundDB
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresMenuRepository) GetAll(ctx context.Context) ([]Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+menuColumns+`
		FROM menus
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY sort ASC, id ASC`,
		platformTenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var m Menu
		if err := scanMenu(rows, &m); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, rows.Err()
}

func (r *PostgresMenuRepository) Create(ctx context.Context, req CreateRepoReq) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m Menu
	err = scanMenu(q.QueryRow(ctx, `
		INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+menuColumns,
		platformTenantID, req.Code, req.Name, req.Subtitle, req.URL, req.Path, req.Icon, req.Sort, req.ParentID, req.Ancestors, req.Visible, req.Enabled,
	), &m)
	if err != nil {
		if strings.Contains(err.Error(), "uk_menu_code") {
			return nil, fmt.Errorf("menu code already exists")
		}
		return nil, fmt.Errorf("create platform menu: %w", err)
	}
	return &m, nil
}

func (r *PostgresMenuRepository) Update(ctx context.Context, id uint, req UpdateRepoReq) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m Menu
	err = scanMenu(q.QueryRow(ctx, `
		UPDATE menus SET
			code = $2, name = $3, subtitle = $4, url = $5, path = $6, icon = $7,
			sort = $8, parent_id = $9, ancestors = $10,
			visible = $11, enabled = $12, updated_at = NOW()
		WHERE is_deleted = FALSE AND tenant_id = $1 AND id = $13
		RETURNING `+menuColumns,
		platformTenantID, req.Code, req.Name, req.Subtitle, req.URL, req.Path, req.Icon,
		req.Sort, req.ParentID, req.Ancestors, req.Visible, req.Enabled, id,
	), &m)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errMenuNotFoundDB
		}
		if strings.Contains(err.Error(), "uk_menu_code") {
			return nil, fmt.Errorf("menu code already exists")
		}
		return nil, fmt.Errorf("update platform menu: %w", err)
	}
	return &m, nil
}

func (r *PostgresMenuRepository) Delete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `
		UPDATE menus SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND tenant_id = $1 AND id = $2`,
		platformTenantID, id)
	if err != nil {
		return fmt.Errorf("delete platform menu: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errMenuNotFoundDB
	}
	return nil
}

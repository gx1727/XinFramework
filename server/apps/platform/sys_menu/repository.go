package sysmenu

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: pool}
}

const sysMenuSelectCols = `id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_at, updated_at`

func scanMenu(row pgx.Row) (*Menu, error) {
	var m Menu
	if err := row.Scan(
		&m.ID, &m.Code, &m.Name, &m.Subtitle, &m.URL, &m.Path, &m.Icon,
		&m.Sort, &m.ParentID, &m.Ancestors, &m.Visible, &m.Enabled,
		&m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysMenuSelectCols+` FROM sys_menus WHERE id = $1 AND is_deleted = FALSE`, id)
	m, err := scanMenu(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysMenuNotFoundDB
		}
		return nil, err
	}
	return m, nil
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysMenuSelectCols+` FROM sys_menus WHERE code = $1 AND is_deleted = FALSE`, code)
	m, err := scanMenu(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysMenuNotFoundDB
		}
		return nil, err
	}
	return m, nil
}

func (r *PostgresRepository) GetAll(ctx context.Context) ([]Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `SELECT `+sysMenuSelectCols+` FROM sys_menus
		WHERE is_deleted = FALSE ORDER BY sort ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Menu{}
	for rows.Next() {
		m, err := scanMenu(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, req CreateRepoReq) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	visible, enabled := req.Visible, req.Enabled
	if !visible {
		visible = true
	}
	if !enabled {
		enabled = true
	}
	row := q.QueryRow(ctx, `INSERT INTO sys_menus
		(code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
		RETURNING `+sysMenuSelectCols,
		req.Code, req.Name, req.Subtitle, req.URL, req.Path, req.Icon,
		req.Sort, req.ParentID, req.Ancestors, visible, enabled, req.CreatedBy)
	return scanMenu(row)
}

func (r *PostgresRepository) Update(ctx context.Context, id uint, req UpdateRepoReq) (*Menu, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	sets := []string{}
	args := []any{}
	idx := 1
	if req.Code != nil {
		sets = append(sets, "code = $"+itoa(idx))
		args = append(args, *req.Code)
		idx++
	}
	if req.Name != nil {
		sets = append(sets, "name = $"+itoa(idx))
		args = append(args, *req.Name)
		idx++
	}
	if req.Subtitle != nil {
		sets = append(sets, "subtitle = $"+itoa(idx))
		args = append(args, *req.Subtitle)
		idx++
	}
	if req.URL != nil {
		sets = append(sets, "url = $"+itoa(idx))
		args = append(args, *req.URL)
		idx++
	}
	if req.Path != nil {
		sets = append(sets, "path = $"+itoa(idx))
		args = append(args, *req.Path)
		idx++
	}
	if req.Icon != nil {
		sets = append(sets, "icon = $"+itoa(idx))
		args = append(args, *req.Icon)
		idx++
	}
	if req.Sort != nil {
		sets = append(sets, "sort = $"+itoa(idx))
		args = append(args, *req.Sort)
		idx++
	}
	if req.ParentID != nil {
		sets = append(sets, "parent_id = $"+itoa(idx))
		args = append(args, *req.ParentID)
		idx++
	}
	if req.Ancestors != nil {
		sets = append(sets, "ancestors = $"+itoa(idx))
		args = append(args, *req.Ancestors)
		idx++
	}
	if req.Visible != nil {
		sets = append(sets, "visible = $"+itoa(idx))
		args = append(args, *req.Visible)
		idx++
	}
	if req.Enabled != nil {
		sets = append(sets, "enabled = $"+itoa(idx))
		args = append(args, *req.Enabled)
		idx++
	}
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}
	sets = append(sets, "updated_by = $"+itoa(idx), "updated_at = NOW()")
	args = append(args, req.UpdatedBy, id)
	row := q.QueryRow(ctx, `UPDATE sys_menus SET `+strings.Join(sets, ", ")+`
		WHERE id = $`+itoa(idx+1)+` AND is_deleted = FALSE
		RETURNING `+sysMenuSelectCols, args...)
	m, err := scanMenu(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysMenuNotFoundDB
		}
		return nil, err
	}
	return m, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint, updatedBy uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	cmd, err := q.Exec(ctx, `UPDATE sys_menus
		SET is_deleted = TRUE, updated_by = $1, updated_at = NOW()
		WHERE id = $2 AND is_deleted = FALSE`, updatedBy, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errSysMenuNotFoundDB
	}
	return nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

package syspermission

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

const sysPermissionSelectCols = `id, menu_id, code, name, action, description, sort, status, created_at, updated_at`

func scanPermission(row pgx.Row) (*Permission, error) {
	var p Permission
	if err := row.Scan(
		&p.ID, &p.MenuID, &p.Code, &p.Name, &p.Action, &p.Description,
		&p.Sort, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*Permission, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysPermissionSelectCols+` FROM sys_permissions WHERE id = $1 AND is_deleted = FALSE`, id)
	p, err := scanPermission(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysPermissionNotFoundDB
		}
		return nil, err
	}
	return p, nil
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) ([]Permission, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `SELECT `+sysPermissionSelectCols+` FROM sys_permissions
		WHERE code = $1 AND is_deleted = FALSE ORDER BY id`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Permission{}
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) List(ctx context.Context, menuID *uint, keyword string, page, size int) ([]Permission, int64, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}
	conds := []string{"is_deleted = FALSE"}
	args := []any{}
	idx := 1
	if menuID != nil {
		conds = append(conds, "menu_id = $"+itoa(idx))
		args = append(args, *menuID)
		idx++
	}
	if keyword != "" {
		conds = append(conds, "(code ILIKE $"+itoa(idx)+" OR name ILIKE $"+itoa(idx)+")")
		args = append(args, "%"+keyword+"%")
		idx++
	}
	whereClause := strings.Join(conds, " AND ")

	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM sys_permissions WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}
	offset := (page - 1) * size
	listArgs := append(append([]any{}, args...), size, offset)
	rows, err := q.Query(ctx, `SELECT `+sysPermissionSelectCols+` FROM sys_permissions
		WHERE `+whereClause+` ORDER BY sort ASC, id ASC
		LIMIT $`+itoa(idx)+` OFFSET $`+itoa(idx+1), listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]Permission, 0, size)
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *p)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, req CreateRepoReq) (*Permission, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	row := q.QueryRow(ctx, `INSERT INTO sys_permissions
		(menu_id, code, name, action, description, sort, status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING `+sysPermissionSelectCols,
		req.MenuID, req.Code, req.Name, req.Action, req.Description, req.Sort, status, req.CreatedBy)
	return scanPermission(row)
}

func (r *PostgresRepository) Update(ctx context.Context, id uint, req UpdateRepoReq) (*Permission, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	sets := []string{}
	args := []any{}
	idx := 1
	if req.MenuID != nil {
		sets = append(sets, "menu_id = $"+itoa(idx))
		args = append(args, *req.MenuID)
		idx++
	}
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
	if req.Action != nil {
		sets = append(sets, "action = $"+itoa(idx))
		args = append(args, *req.Action)
		idx++
	}
	if req.Description != nil {
		sets = append(sets, "description = $"+itoa(idx))
		args = append(args, *req.Description)
		idx++
	}
	if req.Sort != nil {
		sets = append(sets, "sort = $"+itoa(idx))
		args = append(args, *req.Sort)
		idx++
	}
	if req.Status != nil {
		sets = append(sets, "status = $"+itoa(idx))
		args = append(args, *req.Status)
		idx++
	}
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}
	sets = append(sets, "updated_by = $"+itoa(idx), "updated_at = NOW()")
	args = append(args, req.UpdatedBy, id)
	row := q.QueryRow(ctx, `UPDATE sys_permissions SET `+strings.Join(sets, ", ")+`
		WHERE id = $`+itoa(idx+1)+` AND is_deleted = FALSE
		RETURNING `+sysPermissionSelectCols, args...)
	p, err := scanPermission(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysPermissionNotFoundDB
		}
		return nil, err
	}
	return p, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint, updatedBy uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	cmd, err := q.Exec(ctx, `UPDATE sys_permissions
		SET is_deleted = TRUE, updated_by = $1, updated_at = NOW()
		WHERE id = $2 AND is_deleted = FALSE`, updatedBy, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errSysPermissionNotFoundDB
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

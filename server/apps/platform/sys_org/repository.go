package sysorg

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

const sysOrgSelectCols = `id, parent_id, code, name, type, description, admin_code, ancestors, sort, status, created_at, updated_at`

func scanOrg(row pgx.Row) (*Org, error) {
	var o Org
	if err := row.Scan(
		&o.ID, &o.ParentID, &o.Code, &o.Name, &o.Type, &o.Description,
		&o.AdminCode, &o.Ancestors, &o.Sort, &o.Status,
		&o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*Org, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysOrgSelectCols+` FROM sys_orgs WHERE id = $1 AND is_deleted = FALSE`, id)
	o, err := scanOrg(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysOrgNotFoundDB
		}
		return nil, err
	}
	return o, nil
}

func (r *PostgresRepository) List(ctx context.Context, keyword string, page, size int) ([]Org, int64, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}
	conds := []string{"is_deleted = FALSE"}
	args := []any{}
	if keyword != "" {
		conds = append(conds, "(code ILIKE $1 OR name ILIKE $1 OR description ILIKE $1)")
		args = append(args, "%"+keyword+"%")
	}
	whereClause := strings.Join(conds, " AND ")
	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM sys_orgs WHERE `+whereClause, args...).Scan(&total); err != nil {
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
	rows, err := q.Query(ctx, `SELECT `+sysOrgSelectCols+` FROM sys_orgs
		WHERE `+whereClause+` ORDER BY sort ASC, id ASC
		LIMIT $`+itoa(len(args)+1)+` OFFSET $`+itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]Org, 0, size)
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *o)
	}
	return out, total, rows.Err()
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

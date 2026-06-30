package sysuser

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

// PostgresRepository 是 Repository 的 PostgreSQL 实现。
//
// 关键不变量：
//  1. 所有 SQL 走 db.RunInPlatformTx 上下文（bypass_rls=on）。
//  2. sys_users 表**不携带 tenant_id**——platform 域单租户。
//  3. 软删除统一 is_deleted = FALSE 谓词。
type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: pool}
}

// sysUserSelectCols 把 schema 里允许 NULL 的字符串列（code/real_name/nickname/avatar）
// 用 COALESCE 兜底成空串，避免 pgx Scan 时报 "cannot scan NULL into *string"。
// org_id 已在 Go 端用 *uint 指针承载 NULL，无需 COALESCE。
// COALESCE(x, ”) 在 PG 里默认列名是 "coalesce"，用 AS 保留原名以便 Go 端按列名映射。
const sysUserSelectCols = `id, account_id, org_id, COALESCE(code, '') AS code, COALESCE(real_name, '') AS real_name, COALESCE(nickname, '') AS nickname, COALESCE(avatar, '') AS avatar, status, created_at, updated_at`

func scanUser(row pgx.Row) (*User, error) {
	var u User
	if err := row.Scan(
		&u.ID, &u.AccountID, &u.OrgID, &u.Code, &u.RealName, &u.Nickname,
		&u.Avatar, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysUserSelectCols+`
		FROM sys_users WHERE id = $1 AND is_deleted = FALSE`, id)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysUserNotFoundDB
		}
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) GetByAccountID(ctx context.Context, accountID uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysUserSelectCols+`
		FROM sys_users WHERE account_id = $1 AND is_deleted = FALSE`, accountID)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysUserNotFoundDB
		}
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysUserSelectCols+`
		FROM sys_users WHERE code = $1 AND is_deleted = FALSE`, code)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysUserNotFoundDB
		}
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) List(ctx context.Context, keyword string, page, size int) ([]User, int64, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}
	conds := []string{"is_deleted = FALSE"}
	args := []any{}
	if keyword != "" {
		conds = append(conds, "(code ILIKE $1 OR real_name ILIKE $1 OR nickname ILIKE $1)")
		args = append(args, "%"+keyword+"%")
	}
	whereClause := strings.Join(conds, " AND ")

	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM sys_users WHERE `+whereClause, args...).Scan(&total); err != nil {
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
	rows, err := q.Query(ctx, `SELECT `+sysUserSelectCols+`
		FROM sys_users
		WHERE `+whereClause+`
		ORDER BY id DESC
		LIMIT $`+itoa(len(args)+1)+` OFFSET $`+itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]User, 0, size)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *u)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, req CreateRepoReq) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	row := q.QueryRow(ctx, `INSERT INTO sys_users
		(account_id, code, org_id, real_name, nickname, avatar, status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING `+sysUserSelectCols,
		req.AccountID, req.Code, req.OrgID, req.RealName, req.Nickname, req.Avatar, status, req.CreatedBy)
	return scanUser(row)
}

func (r *PostgresRepository) Update(ctx context.Context, id uint, req UpdateRepoReq) (*User, error) {
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
	if req.OrgID != nil {
		sets = append(sets, "org_id = $"+itoa(idx))
		args = append(args, *req.OrgID)
		idx++
	}
	if req.RealName != nil {
		sets = append(sets, "real_name = $"+itoa(idx))
		args = append(args, *req.RealName)
		idx++
	}
	if req.Nickname != nil {
		sets = append(sets, "nickname = $"+itoa(idx))
		args = append(args, *req.Nickname)
		idx++
	}
	if req.Avatar != nil {
		sets = append(sets, "avatar = $"+itoa(idx))
		args = append(args, *req.Avatar)
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

	row := q.QueryRow(ctx, `UPDATE sys_users SET `+strings.Join(sets, ", ")+`
		WHERE id = $`+itoa(idx+1)+` AND is_deleted = FALSE
		RETURNING `+sysUserSelectCols, args...)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysUserNotFoundDB
		}
		return nil, err
	}
	return u, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uint, status int8, updatedBy uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	cmd, err := q.Exec(ctx, `UPDATE sys_users
		SET status = $1, updated_by = $2, updated_at = NOW()
		WHERE id = $3 AND is_deleted = FALSE`,
		status, updatedBy, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errSysUserNotFoundDB
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint, updatedBy uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	cmd, err := q.Exec(ctx, `UPDATE sys_users
		SET is_deleted = TRUE, updated_by = $1, updated_at = NOW()
		WHERE id = $2 AND is_deleted = FALSE`,
		updatedBy, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errSysUserNotFoundDB
	}
	return nil
}

func (r *PostgresRepository) ListRoles(ctx context.Context, userID uint) ([]RoleLite, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT r.id, r.code, r.name
		FROM sys_user_roles ur
		JOIN sys_roles r ON r.id = ur.role_id AND r.is_deleted = FALSE
		WHERE ur.user_id = $1 AND ur.is_deleted = FALSE
		ORDER BY r.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RoleLite{}
	for rows.Next() {
		var rl RoleLite
		if err := rows.Scan(&rl.ID, &rl.Code, &rl.Name); err != nil {
			return nil, err
		}
		out = append(out, rl)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) GrantRole(ctx context.Context, userID, roleID uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `INSERT INTO sys_user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) WHERE is_deleted = FALSE DO NOTHING`,
		userID, roleID)
	return err
}

func (r *PostgresRepository) RevokeRole(ctx context.Context, userID, roleID uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `UPDATE sys_user_roles
		SET is_deleted = TRUE
		WHERE user_id = $1 AND role_id = $2 AND is_deleted = FALSE`,
		userID, roleID)
	return err
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

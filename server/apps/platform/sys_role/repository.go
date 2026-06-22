package sysrole

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

const sysRoleSelectCols = `id, org_id, code, name, description, data_scope, is_default, sort, status, created_at, updated_at`

func scanRole(row pgx.Row) (*Role, error) {
	var r Role
	if err := row.Scan(
		&r.ID, &r.OrgID, &r.Code, &r.Name, &r.Description, &r.DataScope,
		&r.IsDefault, &r.Sort, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*Role, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysRoleSelectCols+` FROM sys_roles WHERE id = $1 AND is_deleted = FALSE`, id)
	role, err := scanRole(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysRoleNotFoundDB
		}
		return nil, err
	}
	return role, nil
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) (*Role, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `SELECT `+sysRoleSelectCols+` FROM sys_roles WHERE code = $1 AND is_deleted = FALSE`, code)
	role, err := scanRole(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysRoleNotFoundDB
		}
		return nil, err
	}
	return role, nil
}

func (r *PostgresRepository) List(ctx context.Context, keyword string, page, size int) ([]Role, int64, error) {
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
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM sys_roles WHERE `+whereClause, args...).Scan(&total); err != nil {
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
	rows, err := q.Query(ctx, `SELECT `+sysRoleSelectCols+` FROM sys_roles
		WHERE `+whereClause+` ORDER BY sort ASC, id ASC
		LIMIT $`+itoa(len(args)+1)+` OFFSET $`+itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]Role, 0, size)
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *role)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, req CreateRepoReq) (*Role, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	row := q.QueryRow(ctx, `INSERT INTO sys_roles
		(org_id, code, name, description, data_scope, is_default, sort, status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING `+sysRoleSelectCols,
		req.OrgID, req.Code, req.Name, req.Description, req.DataScope,
		req.IsDefault, req.Sort, status, req.CreatedBy)
	return scanRole(row)
}

func (r *PostgresRepository) Update(ctx context.Context, id uint, req UpdateRepoReq) (*Role, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	sets := []string{}
	args := []any{}
	idx := 1
	if req.OrgID != nil {
		sets = append(sets, "org_id = $"+itoa(idx))
		args = append(args, *req.OrgID)
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
	if req.Description != nil {
		sets = append(sets, "description = $"+itoa(idx))
		args = append(args, *req.Description)
		idx++
	}
	if req.DataScope != nil {
		sets = append(sets, "data_scope = $"+itoa(idx))
		args = append(args, *req.DataScope)
		idx++
	}
	if req.IsDefault != nil {
		sets = append(sets, "is_default = $"+itoa(idx))
		args = append(args, *req.IsDefault)
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
	row := q.QueryRow(ctx, `UPDATE sys_roles SET `+strings.Join(sets, ", ")+`
		WHERE id = $`+itoa(idx+1)+` AND is_deleted = FALSE
		RETURNING `+sysRoleSelectCols, args...)
	role, err := scanRole(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errSysRoleNotFoundDB
		}
		return nil, err
	}
	return role, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint, updatedBy uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	cmd, err := q.Exec(ctx, `UPDATE sys_roles
		SET is_deleted = TRUE, updated_by = $1, updated_at = NOW()
		WHERE id = $2 AND is_deleted = FALSE`, updatedBy, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errSysRoleNotFoundDB
	}
	return nil
}

func (r *PostgresRepository) ListUsers(ctx context.Context, roleID uint) ([]uint, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `SELECT user_id FROM sys_user_roles WHERE role_id = $1 AND is_deleted = FALSE`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []uint{}
	for rows.Next() {
		var u uint
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ListMenus(ctx context.Context, roleID uint) ([]MenuLite, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT m.id, m.code, m.name
		FROM sys_role_menus rm
		JOIN sys_menus m ON m.id = rm.menu_id AND m.is_deleted = FALSE
		WHERE rm.role_id = $1 AND rm.is_deleted = FALSE
		ORDER BY m.sort ASC, m.id ASC`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MenuLite{}
	for rows.Next() {
		var ml MenuLite
		if err := rows.Scan(&ml.ID, &ml.Code, &ml.Name); err != nil {
			return nil, err
		}
		out = append(out, ml)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return runAssign(ctx, r.db, "sys_role_menus", "role_id", "menu_id", roleID, menuIDs)
}

func (r *PostgresRepository) ListPermissions(ctx context.Context, roleID uint) ([]PermissionLite, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT p.id, p.code, p.name, p.menu_id
		FROM sys_role_permissions rp
		JOIN sys_permissions p ON p.id = rp.permission_id AND p.is_deleted = FALSE
		WHERE rp.role_id = $1 AND rp.is_deleted = FALSE
		ORDER BY p.id`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PermissionLite{}
	for rows.Next() {
		var pl PermissionLite
		if err := rows.Scan(&pl.ID, &pl.Code, &pl.Name, &pl.MenuID); err != nil {
			return nil, err
		}
		out = append(out, pl)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return runAssign(ctx, r.db, "sys_role_permissions", "role_id", "permission_id", roleID, permissionIDs)
}

// runAssign 全量替换角色-资源关联（先清后插）。
// roleID 是主键侧；childIDs 是被关联侧的 ID 列表。
func runAssign(ctx context.Context, pool *pgxpool.Pool, table, pkCol, childCol string, roleID uint, childIDs []uint) error {
	q, err := db.GetQuerier(ctx, pool)
	if err != nil {
		return err
	}
	if _, err := q.Exec(ctx, `DELETE FROM `+table+` WHERE `+pkCol+` = $1`, roleID); err != nil {
		return err
	}
	for _, id := range childIDs {
		if _, err := q.Exec(ctx, `INSERT INTO `+table+` (`+pkCol+`, `+childCol+`) VALUES ($1, $2)`, roleID, id); err != nil {
			return err
		}
	}
	return nil
}

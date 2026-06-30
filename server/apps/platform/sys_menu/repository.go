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

// ListByUserRoles 返回被分配给 callerRoles 任一角色的菜单。
//
// 路径：sys_users(account_id) -> sys_user_roles -> sys_roles(code = ANY($2)) -> sys_role_menus -> sys_menus。
//
// 多角色取并集并去重（一个账号同时拥有 [test, devops] 时，
// 两个角色下的菜单合并去重返回）。
//
// 软删过滤：sys_user_roles / sys_role_menus / sys_menus 都加上 is_deleted = FALSE，
// 与仓内其他查询保持一致。
//
// ⚠️ CTE 内部 SELECT 必须以 `m.` 给每个列加表别名，因为同时 JOIN 了
// sys_role_menus / sys_roles / sys_user_roles / sys_users，每个表都带 `id` 与 `code`，
// PG 会以 SQLSTATE 42702（"column reference 'id' is ambiguous"）拒绝。
// sysMenuSelectCols 是单表专用，不能直接复用。
//
// ⭐ 祖先补全：用户可能只勾选子菜单（如平台用户 106），未勾父菜单（平台管理 100），
// 写路径只把 106 写入 sys_role_menus。如果原样返回，buildTree 看到 106.parent_id=100
// 但 100 不在结果集里，会把 106 当作孤儿根挂出去，导航栏只剩一个孤零零的“平台用户”。
//
// 解法：用递归 CTE 沿 parent_id 链上溯，把所有缺失的祖先都拉进结果集。
// 为什么不用 sys_menus.ancestors 字段：该字段只在 init_seed 中被维护，API 创建/更新
// 的菜单不会自动写入 ancestors（model 透传 req.Ancestors），新菜单 ancestors 经常是
// 空字符串。CTE 走 parent_id 更健壮，且一次查询完成、无 N+1。
//
// 防环：UNION（不是 UNION ALL）天然去重，递归中遇到已访问的节点即停。
const sysMenuSelectColsPrefixed = `m.id, m.code, m.name, m.subtitle, m.url, m.path, m.icon, m.sort, m.parent_id, m.ancestors, m.visible, m.enabled, m.created_at, m.updated_at`

func (r *PostgresRepository) ListByUserRoles(ctx context.Context, accountID uint, callerRoles []string) ([]Menu, error) {
	if len(callerRoles) == 0 {
		return []Menu{}, nil
	}
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	// 递归 CTE：
	//   阶段 1（anchor）= 直接分配给该用户角色的菜单（叶子）
	//   阶段 2（recursive）= 沿 parent_id 链上溯，把所有祖先拉进来
	// 外层 SELECT 从 sys_menus 取完整行（与单表 sysMenuSelectCols 一致，无 JOIN 无歧义）
	rows, err := q.Query(ctx, `
		WITH RECURSIVE menu_chain AS (
			SELECT m.id, m.parent_id
			FROM sys_menus m
			JOIN sys_role_menus rm ON rm.menu_id = m.id AND rm.is_deleted = FALSE
			JOIN sys_roles r ON r.id = rm.role_id AND r.is_deleted = FALSE
			JOIN sys_user_roles sur ON sur.role_id = r.id AND sur.is_deleted = FALSE
			JOIN sys_users su ON su.id = sur.user_id AND su.is_deleted = FALSE
			WHERE m.is_deleted = FALSE
			  AND su.account_id = $1
			  AND r.code = ANY($2)

			UNION

			SELECT m.id, m.parent_id
			FROM sys_menus m
			INNER JOIN menu_chain mc ON m.id = mc.parent_id
			WHERE m.is_deleted = FALSE
		)
		SELECT `+sysMenuSelectCols+`
		FROM sys_menus
		WHERE id IN (SELECT id FROM menu_chain) AND is_deleted = FALSE
		ORDER BY sort ASC, id ASC
	`, accountID, callerRoles)
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

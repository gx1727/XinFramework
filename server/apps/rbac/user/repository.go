package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/permission"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{db: db}
}

// userScopeColumns uses table-qualified column names so the generated
// scope predicates remain unambiguous after the users ⨯ accounts join
// in userFromClause.
var userScopeColumns = permission.ScopeColumns{
	SelfColumn: "u.id",
	OrgID:      "u.org_id",
}

// userSelectColumns is the canonical column list for every user read path.
// phone/email come from the JOINed accounts row; nickname lives on users.
const userSelectColumns = `u.id, u.tenant_id, u.account_id, u.code, u.nickname, u.status,
		u.real_name, u.avatar, u.org_id,
		a.phone, a.email,
		u.created_at, u.updated_at,
		o.name AS org_name`

// userFromClause is the canonical FROM + JOIN for every user read path.
// The join keeps accounts.is_deleted out of the result so soft-deleted
// accounts surface phone/email as NULL (handled by the *string nil-checks).
const userFromClause = `FROM users u
	LEFT JOIN accounts a ON a.id = u.account_id AND a.is_deleted = FALSE
	LEFT JOIN organizations o ON o.id = u.org_id AND o.is_deleted = FALSE`

func buildUserScopeFilter(ctx context.Context) (permission.ScopeFilter, error) {
	uc, ok := xincontext.UserContextFrom(ctx)
	if !ok || uc == nil || uc.UserID == 0 {
		return permission.ScopeFilter{}, nil
	}
	return uc.GetDataScopeFilterFor(userScopeColumns)
}

func rebindScopeSQL(sql string, from, to int) string {
	for i := from; i >= 1; i-- {
		sql = strings.ReplaceAll(sql, fmt.Sprintf("$%d", i), fmt.Sprintf("$%d", to+i-1))
	}
	return sql
}

// scanUser scans a single row produced by the userSelectColumns column list
// into a User. Optional columns (nickname, real_name, avatar, phone, email)
// are nullable; the join may leave phone/email NULL when the account is
// soft-deleted.
func scanUser(row pgx.Row) (*User, error) {
	var u User
	var nickname, realName, avatar, phone, email, orgName *string
	var orgID *uint64
	if err := row.Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &nickname, &u.Status,
		&realName, &avatar, &orgID,
		&phone, &email,
		&u.CreatedAt, &u.UpdatedAt,
		&orgName,
	); err != nil {
		return nil, err
	}
	if nickname != nil {
		u.Nickname = *nickname
	}
	if realName != nil {
		u.RealName = *realName
	}
	if avatar != nil {
		u.Avatar = *avatar
	}
	if phone != nil {
		u.Phone = *phone
	}
	if email != nil {
		u.Email = *email
	}
	if orgID != nil {
		v := uint(*orgID)
		u.OrgID = &v
	}
	if orgName != nil {
		u.OrgName = *orgName
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	u, err := scanUser(q.QueryRow(ctx,
		`SELECT `+userSelectColumns+`
		`+userFromClause+`
		WHERE u.is_deleted = FALSE AND u.id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *PostgresUserRepository) GetByIDScoped(ctx context.Context, id uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	filter, err := buildUserScopeFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + userSelectColumns + " " + userFromClause + ` WHERE u.is_deleted = FALSE AND u.id = $1`
	args := []any{id}
	if !filter.IsEmpty() {
		query += fmt.Sprintf(" AND (%s)", filter.SQL)
		args = append(args, filter.Args...)
	}

	u, err := scanUser(q.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// GetByAccount returns the user bound to (tenantID, accountID).
//
// Phase 3 note: this method satisfies pkgrbac.UserRepository. It also
// keeps a legacy name GetByAccountID for in-package callers that don't
// filter by tenant (uses tenantID=0 implicitly meaning "any tenant").
func (r *PostgresUserRepository) GetByAccount(ctx context.Context, tenantID, accountID uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	u, err := scanUser(q.QueryRow(ctx,
		`SELECT `+userSelectColumns+`
		`+userFromClause+`
		WHERE u.is_deleted = FALSE AND u.tenant_id = $1 AND u.account_id = $2`, tenantID, accountID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// GetByAccountID is a legacy alias used inside the apps/rbac/user
// package. It scopes to no tenant (tenantID=0) for backwards
// compatibility with existing call sites — prefer GetByAccount
// (which scopes by tenantID) for new code.
func (r *PostgresUserRepository) GetByAccountID(ctx context.Context, accountID uint) (*User, error) {
	return r.GetByAccount(ctx, 0, accountID)
}

func (r *PostgresUserRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	u, err := scanUser(q.QueryRow(ctx,
		`SELECT `+userSelectColumns+`
		`+userFromClause+`
		WHERE u.is_deleted = FALSE AND u.tenant_id = $1 AND u.code = $2`, tenantID, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// GetByCodeLegacy is the unscoped variant kept for in-package callers
// that don't filter by tenant. Prefer GetByCode for new code.
func (r *PostgresUserRepository) GetByCodeLegacy(ctx context.Context, code string) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	u, err := scanUser(q.QueryRow(ctx,
		`SELECT `+userSelectColumns+`
		`+userFromClause+`
		WHERE u.is_deleted = FALSE AND u.code = $1`, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error) {
	if tenantID == 0 {
		tenantID, _ = xincontext.TenantIDFrom(ctx)
	}

	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	where, args, argIdx := buildUserListWhere(tenantID, permission.ScopeFilter{}, keyword, nil)

	var total int64
	countSQL := "SELECT COUNT(*) " + userFromClause + " " + where
	if e := q.QueryRow(ctx, countSQL, args...).Scan(&total); e != nil {
		logger.Module("user").Errorf("[List] COUNT failed: %v | sql=%s | args=%v", e, countSQL, args)
		return nil, 0, e
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`SELECT %s %s %s ORDER BY u.id DESC LIMIT $%d OFFSET $%d`,
		userSelectColumns, userFromClause, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list, err := scanUserRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *PostgresUserRepository) ListScoped(ctx context.Context, tenantID uint, keyword string, orgID *uint, page, size int) ([]User, int64, error) {
	if tenantID == 0 {
		tenantID, _ = xincontext.TenantIDFrom(ctx)
	}

	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	filter, err := buildUserScopeFilter(ctx)
	if err != nil {
		return nil, 0, err
	}

	where, args, argIdx := buildUserListWhere(tenantID, filter, keyword, orgID)

	var total int64
	countSQL := "SELECT COUNT(*) " + userFromClause + " " + where
	if e := q.QueryRow(ctx, countSQL, args...).Scan(&total); e != nil {
		logger.Module("user").Errorf("[ListScoped] COUNT failed: %v | sql=%s | args=%v", e, countSQL, args)
		return nil, 0, e
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`SELECT %s %s %s ORDER BY u.id DESC LIMIT $%d OFFSET $%d`,
		userSelectColumns, userFromClause, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list, err := scanUserRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// buildUserListWhere assembles the WHERE clause and args for the user list
// queries. tenantID, the data-scope filter, and the keyword search each
// consume a single $N placeholder (keyword is reused 5x). Returns the
// assembled WHERE, the args slice, and the next free placeholder index
// (used by the caller to append LIMIT/OFFSET).
func buildUserListWhere(tenantID uint, filter permission.ScopeFilter, keyword string, orgID *uint) (string, []any, int) {
	where := "WHERE u.is_deleted = FALSE"
	args := []any{}
	argIdx := 1

	if tenantID > 0 {
		where += fmt.Sprintf(" AND u.tenant_id = $%d", argIdx)
		args = append(args, tenantID)
		argIdx++
	}
	if !filter.IsEmpty() {
		sql := rebindScopeSQL(filter.SQL, len(filter.Args), argIdx)
		where += " AND (" + sql + ")"
		args = append(args, filter.Args...)
		argIdx += len(filter.Args)
	}
	if keyword != "" {
		// 5 spec-required fields: 账号(username) / 昵称(nickname) / 姓名(real_name) / 手机(phone) / 编码(code).
		// One shared $N placeholder reused 5x (PostgreSQL allows this).
		where += fmt.Sprintf(" AND (u.code ILIKE $%d OR u.nickname ILIKE $%d OR u.real_name ILIKE $%d OR a.phone ILIKE $%d OR a.username ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+keyword+"%")
		argIdx++
	}
	if orgID != nil {
		where += fmt.Sprintf(" AND u.org_id = $%d", argIdx)
		args = append(args, *orgID)
		argIdx++
	}
	return where, args, argIdx
}

func scanUserRows(rows pgx.Rows) ([]User, error) {
	var list []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *u)
	}
	return list, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, tenantID, accountID uint, code string, orgID *uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	// Insert on users; org_id is optional (nullable column).
	var orgIDArg interface{}
	if orgID != nil {
		orgIDArg = *orgID
	}

	// RETURNING is the user-row only (no JOIN). org_name gets loaded
	// below by re-reading the canonical JOINed row.
	var insertedID uint
	if err := q.QueryRow(ctx, `
		INSERT INTO users (tenant_id, account_id, code, status, org_id)
		VALUES ($1, $2, $3, 1, $4)
		RETURNING id`,
		tenantID, accountID, code, orgIDArg).Scan(&insertedID); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Re-read via the canonical JOINed path so the returned struct matches
	// the read path used everywhere else.
	u, err := r.GetByID(ctx, insertedID)
	if err != nil {
		return nil, err
	}

	// Fill phone/email from the freshly-created account (preserved behavior).
	if err := q.QueryRow(ctx,
		`SELECT phone, email FROM accounts WHERE id = $1 AND is_deleted = FALSE`,
		accountID).Scan(&u.Phone, &u.Email); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("load account: %w", err)
	}
	return u, nil
}

// Update 全量更新 users 表的 4 个字段（不含 phone/email，跨表见 UpdatePhone）
func (r *PostgresUserRepository) Update(ctx context.Context, id uint, req UpdateUserRepoReq) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var orgIDArg interface{}
	if req.OrgID != nil {
		orgIDArg = *req.OrgID
	}
	tag, err := q.Exec(ctx, `
		UPDATE users SET nickname = $2, real_name = $3, avatar = $4, status = $5, org_id = $6, updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE`,
		id, req.Nickname, req.RealName, req.Avatar, req.Status, orgIDArg)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrUserNotFound
	}

	// 复用 GetByIDScoped 走带 JOIN 的标准读路径，保证 phone/email 正确返回
	return r.GetByIDScoped(ctx, id)
}

// Patch 局部更新：仅修改 req 中非 nil 字段；空 body 等价于 GetByID
func (r *PostgresUserRepository) Patch(ctx context.Context, id uint, req PatchUserRepoReq) (*User, error) {
	sets := make([]string, 0, 4)
	args := make([]interface{}, 0, 5)
	idx := 1

	if req.Nickname != nil {
		sets = append(sets, fmt.Sprintf("nickname = $%d", idx))
		args = append(args, *req.Nickname)
		idx++
	}
	if req.RealName != nil {
		sets = append(sets, fmt.Sprintf("real_name = $%d", idx))
		args = append(args, *req.RealName)
		idx++
	}
	if req.Avatar != nil {
		sets = append(sets, fmt.Sprintf("avatar = $%d", idx))
		args = append(args, *req.Avatar)
		idx++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", idx))
		args = append(args, *req.Status)
		idx++
	}
	if req.OrgID != nil {
		var arg interface{}
		if *req.OrgID == 0 {
			arg = nil // explicit "remove from org"
		} else {
			arg = *req.OrgID
		}
		sets = append(sets, fmt.Sprintf("org_id = $%d", idx))
		args = append(args, arg)
		idx++
	}

	if len(sets) == 0 {
		return r.GetByIDScoped(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d AND is_deleted = FALSE`,
		strings.Join(sets, ", "), idx)

	tag, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("patch user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrUserNotFound
	}

	return r.GetByIDScoped(ctx, id)
}

func (r *PostgresUserRepository) UpdateStatus(ctx context.Context, id uint, status int8) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE users SET status = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE users SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateOrg 调整用户的主组织；orgID 为 nil 或 0 都表示移出组织（org_id 置 NULL）。
// org 是否存在 / 是否同租户的校验由 service 层负责，仓库只做原子写入。
func (r *PostgresUserRepository) UpdateOrg(ctx context.Context, id uint, orgID *uint) (*User, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var arg interface{}
	if orgID != nil && *orgID != 0 {
		arg = *orgID
	}

	tag, err := q.Exec(ctx, `
		UPDATE users SET org_id = $2, updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE`,
		id, arg)
	if err != nil {
		return nil, fmt.Errorf("update user org: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrUserNotFoundDB
	}
	return r.GetByIDScoped(ctx, id)
}

// UpdatePhone writes phone to the user's underlying account row (phone is
// an account-level attribute, not a per-tenant user attribute). The single
// UPDATE...FROM statement resolves account_id without a separate lookup.
func (r *PostgresUserRepository) UpdatePhone(ctx context.Context, userID uint, phone string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE accounts a SET phone = $2, updated_at = NOW()
		FROM users u
		WHERE u.id = $1
		  AND a.id = u.account_id
		  AND u.is_deleted = FALSE
		  AND a.is_deleted = FALSE`, userID, phone)
	if err != nil {
		return fmt.Errorf("update user phone: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, id uint, nickname, avatar string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE users SET nickname = $2, avatar = $3, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id, nickname, avatar)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpdateAvatar(ctx context.Context, id uint, avatar string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE users SET avatar = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id, avatar)
	if err != nil {
		return fmt.Errorf("update user avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

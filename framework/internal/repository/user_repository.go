package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/model"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) model.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
		&u.RealName, &u.Avatar, &u.Phone, &u.Email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByAccountID(ctx context.Context, accountID uint) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE is_deleted = FALSE AND account_id = $1`, accountID).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
		&u.RealName, &u.Avatar, &u.Phone, &u.Email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByCode(ctx context.Context, code string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE is_deleted = FALSE AND code = $1`, code).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
		&u.RealName, &u.Avatar, &u.Phone, &u.Email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]model.User, int64, error) {
	where := "WHERE is_deleted = FALSE"
	args := []interface{}{}
	argIdx := 1

	if tenantID > 0 {
		where += fmt.Sprintf(" AND tenant_id = $%d", argIdx)
		args = append(args, tenantID)
		argIdx++
	}
	if keyword != "" {
		where += fmt.Sprintf(" AND (code ILIKE $%d OR nickname ILIKE $%d OR real_name ILIKE $%d OR phone ILIKE $%d)", argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
			&u.RealName, &u.Avatar, &u.Phone, &u.Email,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, u)
	}
	return list, total, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, tenantID, accountID uint, code string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (tenant_id, account_id, code, status)
		VALUES ($1, $2, $3, 1)
		RETURNING id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at`,
		tenantID, accountID, code).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
		&u.RealName, &u.Avatar, &u.Phone, &u.Email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepository) UpdateStatus(ctx context.Context, id uint, status int8) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET status = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uint) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpdatePhone(ctx context.Context, userID uint, phone string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET phone = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, userID, phone)
	if err != nil {
		return fmt.Errorf("update user phone: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

// PostgresRoleRepository implements model.RoleRepository
type PostgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) model.RoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
		FROM roles
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrRoleNotFound
		}
		return nil, err
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*model.Role, error) {
	var role model.Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
		FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrRoleNotFound
		}
		return nil, err
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetUserRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.tenant_id, r.org_id, r.code, r.name, r.description, r.data_scope, r.extend, r.is_default, r.sort, r.status, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.is_deleted = FALSE`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		var extend []byte
		if err := rows.Scan(
			&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
			&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if extend != nil {
			role.Extend = string(extend)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]model.Role, int64, error) {
	offset := (page - 1) * size
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
		FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1
		AND ($2 = '' OR (code ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%'))
		ORDER BY sort ASC, id ASC
		LIMIT $3 OFFSET $4`, tenantID, keyword, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		var extend []byte
		if err := rows.Scan(
			&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
			&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if extend != nil {
			role.Extend = string(extend)
		}
		roles = append(roles, role)
	}

	var total int64
	r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1
		AND ($2 = '' OR (code ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%'))`,
		tenantID, keyword).Scan(&total)

	return roles, total, nil
}

func (r *PostgresRoleRepository) Create(ctx context.Context, tenantID uint, req model.CreateRoleRepoReq) (*model.Role, error) {
	var role model.Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		INSERT INTO roles (tenant_id, code, name, description, data_scope, is_default, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
	`, tenantID, req.Code, req.Name, req.Description, req.DataScope, req.IsDefault, req.Sort, req.Status).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) Update(ctx context.Context, id uint, req model.UpdateRoleRepoReq) (*model.Role, error) {
	var role model.Role
	var extend []byte
	err := r.db.QueryRow(ctx, `
		UPDATE roles SET name = $1, description = $2, data_scope = $3, is_default = $4, sort = $5, status = $6, updated_at = NOW()
		WHERE id = $7 AND is_deleted = FALSE
		RETURNING id, tenant_id, org_id, code, name, description, data_scope, extend, is_default, sort, status, created_at, updated_at
	`, req.Name, req.Description, req.DataScope, req.IsDefault, req.Sort, req.Status, id).Scan(
		&role.ID, &role.TenantID, &role.OrgID, &role.Code, &role.Name, &role.Description,
		&role.DataScope, &extend, &role.IsDefault, &role.Sort, &role.Status, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrRoleNotFound
		}
		return nil, fmt.Errorf("update role: %w", err)
	}
	if extend != nil {
		role.Extend = string(extend)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id uint) error {
	_, err := r.db.Exec(ctx, `UPDATE roles SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

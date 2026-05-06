package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uint) (_ *User, err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	var u User
	var nickname, realName, avatar, phone, email *string
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &nickname, &u.Status,
		&realName, &avatar, &phone, &email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
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
	return &u, nil
}

func (r *PostgresUserRepository) GetByAccountID(ctx context.Context, accountID uint) (_ *User, err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	var u User
	var nickname, realName, avatar, phone, email *string
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE is_deleted = FALSE AND account_id = $1`, accountID).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &nickname, &u.Status,
		&realName, &avatar, &phone, &email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
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
	return &u, nil
}

func (r *PostgresUserRepository) GetByCode(ctx context.Context, code string) (_ *User, err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	var u User
	var nickname, realName, avatar, phone, email *string
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE is_deleted = FALSE AND code = $1`, code).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &nickname, &u.Status,
		&realName, &avatar, &phone, &email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
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
	return &u, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) (_ []User, _ int64, err error) {
	if tenantID == 0 {
		tenantID, _ = xincontext.TenantIDFrom(ctx)
	}

	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

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
	err = q.QueryRow(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total)
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

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []User
	for rows.Next() {
		var u User
		var nickname, realName, avatar, phone, email *string
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.AccountID, &u.Code, &nickname, &u.Status,
			&realName, &avatar, &phone, &email,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
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
		list = append(list, u)
	}
	return list, total, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, tenantID, accountID uint, code string) (_ *User, err error) {
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	var u User
	var nickname, realName, avatar, phone, email *string
	err = q.QueryRow(ctx, `
		INSERT INTO users (tenant_id, account_id, code, status)
		VALUES ($1, $2, $3, 1)
		RETURNING id, tenant_id, account_id, code, nickname, status, real_name, avatar, phone, email, created_at, updated_at`,
		tenantID, accountID, code).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &nickname, &u.Status,
		&realName, &avatar, &phone, &email,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
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
	return &u, nil
}

func (r *PostgresUserRepository) UpdateStatus(ctx context.Context, id uint, status int8) (err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

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

func (r *PostgresUserRepository) Delete(ctx context.Context, id uint) (err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

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

func (r *PostgresUserRepository) UpdatePhone(ctx context.Context, userID uint, phone string) (err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	tag, err := q.Exec(ctx, `
		UPDATE users SET phone = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, userID, phone)
	if err != nil {
		return fmt.Errorf("update user phone: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, id uint, nickname, avatar string) (err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

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

func (r *PostgresUserRepository) UpdateAvatar(ctx context.Context, id uint, avatar string) (err error) {
	tenantID, _ := xincontext.TenantIDFrom(ctx)
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

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

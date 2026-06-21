package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// PostgresAccountRepository implements AccountRepository
type PostgresAccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) AccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) GetByID(ctx context.Context, id uint) (*Account, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var a Account
	err = q.QueryRow(ctx, `
		SELECT id, username, phone, email, real_name, status, created_at, updated_at
		FROM accounts
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&a.ID, &a.Username, &a.Phone, &a.Email, &a.RealName,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresAccountRepository) GetByUsername(ctx context.Context, username string) (*Account, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var a Account
	err = q.QueryRow(ctx, `
		SELECT id, username, phone, email, real_name, status, created_at, updated_at
		FROM accounts
		WHERE is_deleted = FALSE AND username = $1`, username).Scan(
		&a.ID, &a.Username, &a.Phone, &a.Email, &a.RealName,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresAccountRepository) GetByPhone(ctx context.Context, phone string) (*Account, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var a Account
	err = q.QueryRow(ctx, `
		SELECT id, username, phone, email, real_name, status, created_at, updated_at
		FROM accounts
		WHERE is_deleted = FALSE AND phone = $1`, phone).Scan(
		&a.ID, &a.Username, &a.Phone, &a.Email, &a.RealName,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresAccountRepository) GetByEmail(ctx context.Context, email string) (*Account, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var a Account
	err = q.QueryRow(ctx, `
		SELECT id, username, phone, email, real_name, status, created_at, updated_at
		FROM accounts
		WHERE is_deleted = FALSE AND email = $1`, email).Scan(
		&a.ID, &a.Username, &a.Phone, &a.Email, &a.RealName,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresAccountRepository) Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*Account, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var a Account
	err = q.QueryRow(ctx, `
		INSERT INTO accounts (username, phone, email, real_name, password)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, phone, email, real_name, status, created_at, updated_at`,
		username, phone, email, realName, passwordHash).Scan(
		&a.ID, &a.Username, &a.Phone, &a.Email, &a.RealName,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return &a, nil
}

func (r *PostgresAccountRepository) Exists(ctx context.Context, account string) (bool, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return false, err
	}

	var exists bool
	err = q.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM accounts
			WHERE is_deleted = FALSE
			AND (phone = $1 OR email = $1 OR username = $1)
		)`, account).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetPasswordAndStatus 取账号的 password_hash + id + status。
// 用于 platform-login：账号可能未绑 user 行（无 tenant），所以走 accounts 表直接验证。
//
// account 字段：username / phone / email 任一即可，按优先级匹配（username 最先）。
// 返回 (passwordHash, accountID, status, err)：
//   - passwordHash 用于 verifyPassword
//   - accountID 用于查 platform_roles
//   - status 必须 == 1 才能登录
func (r *PostgresAccountRepository) GetPasswordAndStatus(ctx context.Context, account string) (string, uint, int8, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return "", 0, 0, err
	}

	var passwordHash string
	var id uint
	var status int8
	err = q.QueryRow(ctx, `
		SELECT id, password, status FROM accounts
		WHERE is_deleted = FALSE
		AND (username = $1 OR phone = $1 OR email = $1)
		ORDER BY (username = $1) DESC
		LIMIT 1
	`, account).Scan(&id, &passwordHash, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, 0, errAccountNotFound
		}
		return "", 0, 0, err
	}
	return passwordHash, id, status, nil
}

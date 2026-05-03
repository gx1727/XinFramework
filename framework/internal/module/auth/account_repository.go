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
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

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
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

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
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

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
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

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
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

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
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return false, err
	}
	defer release()

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

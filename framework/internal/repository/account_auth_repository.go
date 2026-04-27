package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/model"
)

// PostgresAccountAuthRepository implements model.AccountAuthRepository
type PostgresAccountAuthRepository struct {
	db *pgxpool.Pool
}

func NewAccountAuthRepository(db *pgxpool.Pool) model.AccountAuthRepository {
	return &PostgresAccountAuthRepository{db: db}
}

func (r *PostgresAccountAuthRepository) GetByOpenID(ctx context.Context, tenantID uint, authType, openID string) (*model.AccountAuth, error) {
	var a model.AccountAuth
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, type, openid, unionid, nickname, avatar, session_key, created_at, updated_at
		FROM account_auths
		WHERE is_deleted = FALSE AND tenant_id = $1 AND type = $2 AND openid = $3`, tenantID, authType, openID).Scan(
		&a.ID, &a.TenantID, &a.AccountID, &a.Type, &a.OpenID, &a.UnionID,
		&a.Nickname, &a.Avatar, &a.SessionKey, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrAccountAuthNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresAccountAuthRepository) GetByAccountID(ctx context.Context, accountID uint) ([]model.AccountAuth, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, account_id, type, openid, unionid, nickname, avatar, session_key, created_at, updated_at
		FROM account_auths
		WHERE is_deleted = FALSE AND account_id = $1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.AccountAuth
	for rows.Next() {
		var a model.AccountAuth
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.AccountID, &a.Type, &a.OpenID, &a.UnionID,
			&a.Nickname, &a.Avatar, &a.SessionKey, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *PostgresAccountAuthRepository) Create(ctx context.Context, tenantID, accountID uint, authType, openID, unionID, sessionKey string) (*model.AccountAuth, error) {
	var a model.AccountAuth
	err := r.db.QueryRow(ctx, `
		INSERT INTO account_auths (tenant_id, account_id, type, openid, unionid, session_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, account_id, type, openid, unionid, nickname, avatar, session_key, created_at, updated_at`,
		tenantID, accountID, authType, openID, unionID, sessionKey).Scan(
		&a.ID, &a.TenantID, &a.AccountID, &a.Type, &a.OpenID, &a.UnionID,
		&a.Nickname, &a.Avatar, &a.SessionKey, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create account auth: %w", err)
	}
	return &a, nil
}

func (r *PostgresAccountAuthRepository) UpdateSessionKey(ctx context.Context, id uint, sessionKey string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE account_auths SET session_key = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id, sessionKey)
	if err != nil {
		return fmt.Errorf("update session key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrAccountAuthNotFound
	}
	return nil
}

func (r *PostgresAccountAuthRepository) Delete(ctx context.Context, id uint) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE account_auths SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete account auth: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrAccountAuthNotFound
	}
	return nil
}

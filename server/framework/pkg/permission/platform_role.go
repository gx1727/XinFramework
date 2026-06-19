package permission

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// PlatformRoleRepository 平台级角色访问接口（跨租户）
type PlatformRoleRepository interface {
	// GetRolesByAccountID 返回账号拥有的所有平台级角色
	GetRolesByAccountID(ctx context.Context, accountID uint) ([]string, error)
	// GetRolesByUserID 通过 user_id 反查 account_id 再取平台角色
	GetRolesByUserID(ctx context.Context, userID uint) ([]string, error)
	// Grant 给账号添加平台角色（幂等）
	Grant(ctx context.Context, accountID uint, role string) error
	// Revoke 撤销账号的平台角色
	Revoke(ctx context.Context, accountID uint, role string) error
}

// PostgresPlatformRoleRepository 基于 account_roles 表的实现
type PostgresPlatformRoleRepository struct {
	db *pgxpool.Pool
}

func NewPlatformRoleRepository(pool *pgxpool.Pool) *PostgresPlatformRoleRepository {
	return &PostgresPlatformRoleRepository{db: pool}
}

func (r *PostgresPlatformRoleRepository) GetRolesByAccountID(ctx context.Context, accountID uint) ([]string, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT role FROM account_roles WHERE account_id = $1
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("get platform roles: %w", err)
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresPlatformRoleRepository) GetRolesByUserID(ctx context.Context, userID uint) ([]string, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var accountID uint
	err = q.QueryRow(ctx, `
		SELECT account_id FROM users WHERE id = $1 AND is_deleted = FALSE
	`, userID).Scan(&accountID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user account_id: %w", err)
	}
	if accountID == 0 {
		return nil, nil
	}
	return r.GetRolesByAccountID(ctx, accountID)
}

func (r *PostgresPlatformRoleRepository) Grant(ctx context.Context, accountID uint, role string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		INSERT INTO account_roles (account_id, role)
		VALUES ($1, $2)
		ON CONFLICT (account_id, role) DO NOTHING
	`, accountID, role)
	if err != nil {
		return fmt.Errorf("grant platform role: %w", err)
	}
	return nil
}

func (r *PostgresPlatformRoleRepository) Revoke(ctx context.Context, accountID uint, role string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `DELETE FROM account_roles WHERE account_id = $1 AND role = $2`, accountID, role)
	if err != nil {
		return fmt.Errorf("revoke platform role: %w", err)
	}
	return nil
}

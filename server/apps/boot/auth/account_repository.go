package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
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

// GetAccountIDByUserID 通过 user_id 反查 account_id。
//
// 用于 Refresh 切租户流程：从 refresh token 的 claims.UserID（= users.id）
// 反查该 user 所属的 account_id，再用 ListTenantIdentities 跨租户列身份。
//
// RLS 说明：本方法查 users 表（启用 RLS），调用方需保证 ctx 处于
// 能查到该 user 行的租户上下文（典型场景：中间件已注入 claims.TenantID 的事务）。
// 在租户事务外调用会被 RLS 拒绝（这是有意的：防越权反查）。
//
// 返回 ErrAccountNotFound 如果 userID 不存在或被软删。
func (r *PostgresAccountRepository) GetAccountIDByUserID(ctx context.Context, userID uint) (uint, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return 0, err
	}

	var accountID uint
	err = q.QueryRow(ctx, `
		SELECT account_id FROM users
		WHERE id = $1 AND is_deleted = FALSE
		LIMIT 1
	`, userID).Scan(&accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrAccountNotFound
		}
		return 0, fmt.Errorf("get account_id by user_id: %w", err)
	}
	return accountID, nil
}

// ListTenantIdentities 列出账号在所有租户的用户身份。
//
// 路径 B 下 accounts 不绑 tenant，但 users 表 (account_id, tenant_id) 复合唯一
// 索引允许一账号对应多租户用户。该方法返回账号在所有启用租户里的 users 身份，
// 用于 LoginPrecheck 让前端选择登录身份。
//
// RLS：跨租户查 users / tenants / user_roles（均启用了 RLS），走
// db.RunInPlatformTx 开启 app.bypass_rls='on' 绕过。
//
// 返回空切片（不是 nil）如果账号没有 tenant 身份。调用方负责区分"无身份"和"出错"。
func (r *PostgresAccountRepository) ListTenantIdentities(ctx context.Context, accountID uint) ([]pkgauth.TenantIdentity, error) {
	var identities []pkgauth.TenantIdentity

	err := db.RunInPlatformTx(ctx, r.db, func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx, r.db)
		if err != nil {
			return err
		}

		rows, err := q.Query(ctx, `
			SELECT
				t.id, t.code, t.name,
				u.id, u.code,
				COALESCE(u.nickname, ''), COALESCE(u.real_name, ''), COALESCE(u.avatar, ''),
				COALESCE(a.email, ''),
				COALESCE((
					SELECT r.code FROM user_roles ur
					JOIN roles r ON r.id = ur.role_id
					WHERE ur.user_id = u.id AND ur.is_deleted = FALSE AND r.is_deleted = FALSE
					ORDER BY ur.id ASC LIMIT 1
				), 'user') AS role_code
			FROM users u
			JOIN accounts a ON a.id = u.account_id
			JOIN tenants t ON t.id = u.tenant_id
			WHERE u.account_id = $1
			  AND u.is_deleted = FALSE
			  AND t.is_deleted = FALSE
			  AND t.status = 1
			ORDER BY t.id ASC
		`, accountID)
		if err != nil {
			return fmt.Errorf("list tenant identities: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ti pkgauth.TenantIdentity
			if err := rows.Scan(
				&ti.TenantID, &ti.TenantCode, &ti.TenantName,
				&ti.UserID, &ti.UserCode,
				&ti.Nickname, &ti.RealName, &ti.Avatar,
				&ti.Email,
				&ti.Role,
			); err != nil {
				return fmt.Errorf("scan tenant identity: %w", err)
			}
			identities = append(identities, ti)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}
	return identities, nil
}

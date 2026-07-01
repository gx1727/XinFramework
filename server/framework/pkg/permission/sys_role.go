package permission

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// SysRoleRepository sys 级角色访问接口（跨租户）
type SysRoleRepository interface {
	// GetRolesByAccountID 返回账号拥有的所有 sys 级角色
	GetRolesByAccountID(ctx context.Context, accountID uint) ([]string, error)
	// GetRolesByUserID 通过 user_id 反查 account_id 再取 sys 角色
	GetRolesByUserID(ctx context.Context, userID uint) ([]string, error)
	// Grant 给账号添加 sys 角色（幂等）
	Grant(ctx context.Context, accountID uint, role string) error
	// Revoke 撤销账号的 sys 角色
	Revoke(ctx context.Context, accountID uint, role string) error
}

// PostgresSysRoleRepository 基于 sys_users / sys_user_roles / sys_roles 的实现。
// 0023.3 终态：account_roles 表已 drop，sys 角色改走 sys_* 三表 join。
type PostgresSysRoleRepository struct {
	db *pgxpool.Pool
}

func NewSysRoleRepository(pool *pgxpool.Pool) *PostgresSysRoleRepository {
	return &PostgresSysRoleRepository{db: pool}
}

// selectRolesForAccountID 是核心 join：account_id -> sys_users -> sys_user_roles -> sys_roles.code。
// 注意：sys_user_roles 与 sys_roles 都有 is_deleted，必须各自加谓词，否则 soft-delete 的关联/角色仍会返回。
const selectRolesForAccountID = `
	SELECT DISTINCT sr.code
	FROM sys_users su
	JOIN sys_user_roles sur ON sur.user_id = su.id AND sur.is_deleted = FALSE
	JOIN sys_roles sr ON sr.id = sur.role_id AND sr.is_deleted = FALSE
	WHERE su.account_id = $1 AND su.is_deleted = FALSE
`

func (r *PostgresSysRoleRepository) GetRolesByAccountID(ctx context.Context, accountID uint) ([]string, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, selectRolesForAccountID, accountID)
	if err != nil {
		return nil, fmt.Errorf("get sys roles: %w", err)
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

func (r *PostgresSysRoleRepository) GetRolesByUserID(ctx context.Context, userID uint) ([]string, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var accountID uint
	err = q.QueryRow(ctx, `
		SELECT account_id FROM tenant_users WHERE id = $1 AND is_deleted = FALSE
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

// Grant 通过 account_id + role.code 找到 sys_users.id 与 sys_roles.id，
// 然后插入 sys_user_roles。account_roles 表已 drop，不再有 account_id + role string 直插。
func (r *PostgresSysRoleRepository) Grant(ctx context.Context, accountID uint, role string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		INSERT INTO sys_user_roles (user_id, role_id)
		SELECT su.id, sr.id
		FROM sys_users su, sys_roles sr
		WHERE su.account_id = $1
		  AND sr.code = $2
		  AND su.is_deleted = FALSE
		  AND sr.is_deleted = FALSE
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, accountID, role)
	if err != nil {
		return fmt.Errorf("grant sys role: %w", err)
	}
	return nil
}

// Revoke 软删：UPDATE is_deleted = TRUE 而非 DELETE，保留审计痕迹。
// 与 sys_user_roles 表的 is_deleted 软删约定一致。
func (r *PostgresSysRoleRepository) Revoke(ctx context.Context, accountID uint, role string) error {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		UPDATE sys_user_roles sur
		SET is_deleted = TRUE, updated_at = NOW()
		FROM sys_users su, sys_roles sr
		WHERE sur.user_id = su.id
		  AND sur.role_id = sr.id
		  AND su.account_id = $1
		  AND sr.code = $2
		  AND sur.is_deleted = FALSE
	`, accountID, role)
	if err != nil {
		return fmt.Errorf("revoke sys role: %w", err)
	}
	return nil
}

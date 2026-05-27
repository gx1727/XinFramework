package role

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// RoleMenuRepository 角色菜单关联数据访问接口
type RoleMenuRepository interface {
	GetByRoleID(ctx context.Context, roleID uint) ([]uint, error) // 返回菜单ID列表
	SetForRole(ctx context.Context, roleID uint, menuIDs []uint) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
}

// PostgresRoleMenuRepository 实现 RoleMenuRepository
type PostgresRoleMenuRepository struct {
	db *pgxpool.Pool
}

func NewRoleMenuRepository(db *pgxpool.Pool) RoleMenuRepository {
	return &PostgresRoleMenuRepository{db: db}
}

func (r *PostgresRoleMenuRepository) GetByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT menu_id FROM role_menus
		WHERE is_deleted = FALSE AND role_id = $1
		ORDER BY menu_id ASC`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuIDs []uint
	for rows.Next() {
		var menuID uint
		if err := rows.Scan(&menuID); err != nil {
			return nil, err
		}
		menuIDs = append(menuIDs, menuID)
	}
	return menuIDs, nil
}

func (r *PostgresRoleMenuRepository) SetForRole(ctx context.Context, roleID uint, menuIDs []uint) error {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}

	// 先软删除现有关联
	_, err = q.Exec(ctx, `
		UPDATE role_menus SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("delete existing role menus: %w", err)
	}

	// 批量插入新关联
	for _, menuID := range menuIDs {
		_, err = q.Exec(ctx, `
			INSERT INTO role_menus (role_id, menu_id)
			VALUES ($1, $2)
			ON CONFLICT (role_id, menu_id) WHERE is_deleted = FALSE
			DO UPDATE SET is_deleted = FALSE, updated_at = NOW()`, roleID, menuID)
		if err != nil {
			return fmt.Errorf("insert role menu: %w", err)
		}
	}
	return nil
}

func (r *PostgresRoleMenuRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		UPDATE role_menus SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND role_id = $1`, roleID)
	return err
}

// RoleMenuResp 角色菜单权限响应
type RoleMenuResp struct {
	MenuIDs []uint `json:"menu_ids"`
}

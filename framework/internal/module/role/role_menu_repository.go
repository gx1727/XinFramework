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

	// 获取当前已有的菜单ID
	existingIDs, err := r.GetByRoleID(ctx, roleID)
	if err != nil {
		return err
	}

	// 转换为 map 便于比较
	existingMap := make(map[uint]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	newMap := make(map[uint]bool)
	for _, id := range menuIDs {
		newMap[id] = true
	}

	// 1. 删除"原来有、现在没有"的关联
	for _, existingID := range existingIDs {
		if !newMap[existingID] {
			_, err = q.Exec(ctx, `
				UPDATE role_menus SET is_deleted = TRUE, updated_at = NOW()
				WHERE is_deleted = FALSE AND role_id = $1 AND menu_id = $2`, roleID, existingID)
			if err != nil {
				return fmt.Errorf("delete role menu: %w", err)
			}
		}
	}

	// 2. 插入"原来没有、现在有"的关联
	for _, newID := range menuIDs {
		if !existingMap[newID] {
			_, err = q.Exec(ctx, `
				INSERT INTO role_menus (role_id, menu_id, tenant_id)
				VALUES ($1, $2, (
					SELECT tenant_id FROM roles WHERE id = $1 AND is_deleted = FALSE
				))
				ON CONFLICT (role_id, menu_id) WHERE is_deleted = FALSE
				DO UPDATE SET is_deleted = FALSE, updated_at = NOW()`, roleID, newID)
			if err != nil {
				return fmt.Errorf("insert role menu: %w", err)
			}
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

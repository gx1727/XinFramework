package permission

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// RoleResourceRepository 角色资源关联数据访问接口
type RoleResourceRepository interface {
	GetByRoleID(ctx context.Context, roleID uint) ([]uint, error) // 返回资源ID列表
	SetForRole(ctx context.Context, roleID uint, resourceIDs []uint) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
}

// PostgresRoleResourceRepository 实现 RoleResourceRepository
type PostgresRoleResourceRepository struct {
	db *pgxpool.Pool
}

func NewRoleResourceRepository(db *pgxpool.Pool) RoleResourceRepository {
	return &PostgresRoleResourceRepository{db: db}
}

func (r *PostgresRoleResourceRepository) GetByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT resource_id FROM role_resources
		WHERE is_deleted = FALSE AND role_id = $1
		ORDER BY resource_id ASC`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resourceIDs []uint
	for rows.Next() {
		var resourceID uint
		if err := rows.Scan(&resourceID); err != nil {
			return nil, err
		}
		resourceIDs = append(resourceIDs, resourceID)
	}
	return resourceIDs, nil
}

func (r *PostgresRoleResourceRepository) SetForRole(ctx context.Context, roleID uint, resourceIDs []uint) error {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}

	// 获取当前已有的资源ID
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
	for _, id := range resourceIDs {
		newMap[id] = true
	}

	// 1. 删除"原来有、现在没有"的关联
	for _, existingID := range existingIDs {
		if !newMap[existingID] {
			_, err = q.Exec(ctx, `
				UPDATE role_resources SET is_deleted = TRUE, updated_at = NOW()
				WHERE is_deleted = FALSE AND role_id = $1 AND resource_id = $2`, roleID, existingID)
			if err != nil {
				return fmt.Errorf("delete role resource: %w", err)
			}
		}
	}

	// 2. 插入"原来没有、现在有"的关联
	for _, newID := range resourceIDs {
		if !existingMap[newID] {
			_, err = q.Exec(ctx, `
				INSERT INTO role_resources (role_id, resource_id, tenant_id)
				VALUES ($1, $2, (
					SELECT tenant_id FROM roles WHERE id = $1 AND is_deleted = FALSE
				))
				ON CONFLICT (role_id, resource_id) WHERE is_deleted = FALSE
				DO UPDATE SET is_deleted = FALSE, updated_at = NOW()`, roleID, newID)
			if err != nil {
				return fmt.Errorf("insert role resource: %w", err)
			}
		}
	}

	return nil
}

func (r *PostgresRoleResourceRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		UPDATE role_resources SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND role_id = $1`, roleID)
	return err
}

// RoleResourceResp 角色资源权限响应
type RoleResourceResp struct {
	ResourceIDs []uint `json:"resource_ids"`
}

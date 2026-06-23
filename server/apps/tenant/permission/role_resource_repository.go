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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}

	if len(resourceIDs) == 0 {
		_, err = q.Exec(ctx, `
			UPDATE role_resources SET is_deleted = TRUE, updated_at = NOW()
			WHERE is_deleted = FALSE AND role_id = $1`, roleID)
		if err != nil {
			return fmt.Errorf("delete all role resources: %w", err)
		}
		return nil
	}

	_, err = q.Exec(ctx, `
		UPDATE role_resources SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND role_id = $1
		  AND resource_id NOT IN (
			SELECT unnest($2::bigint[])
		  )`, roleID, resourceIDs)
	if err != nil {
		return fmt.Errorf("delete removed role resources: %w", err)
	}

	var tenantID int64
	err = q.QueryRow(ctx, `SELECT tenant_id FROM roles WHERE id = $1 AND is_deleted = FALSE`, roleID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("get tenant_id for role: %w", err)
	}

	_, err = q.Exec(ctx, `
		INSERT INTO role_resources (role_id, resource_id, tenant_id)
		SELECT $1, unnest, $3
		FROM unnest($2::bigint[]) AS unnest
		ON CONFLICT (role_id, resource_id) WHERE is_deleted = FALSE
		DO UPDATE SET is_deleted = FALSE, updated_at = NOW()`, roleID, resourceIDs, tenantID)
	if err != nil {
		return fmt.Errorf("insert role resources: %w", err)
	}

	return nil
}
func (r *PostgresRoleResourceRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	q, err := db.GetQuerier(ctx, r.db)
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

package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// PostgresTenantRepository implements TenantRepository
type PostgresTenantRepository struct {
	db *pgxpool.Pool
}

// tenantSelectCols 租户 SELECT 列清单（COALESCE 兜底 NULL，避免 scan 进 *string 失败）。
// 字符串列用 '' 兜底，数值列用 0 兜底，jsonb 用 '{}'::jsonb 兜底。
// 三个 SELECT（GetByID / GetByCode / List）必须保持一致。
const tenantSelectCols = `
	id, code, name, status,
	COALESCE(contact, '') AS contact,
	COALESCE(phone, '')   AS phone,
	COALESCE(email, '')   AS email,
	COALESCE(province, '') AS province,
	COALESCE(city, '')    AS city,
	COALESCE(area, '')    AS area,
	COALESCE(address, '') AS address,
	COALESCE(config, '{}'::jsonb)::text    AS config,
	COALESCE(dashboard, '')              AS dashboard,
	created_at, updated_at,
	COALESCE(created_by, 0) AS created_by,
	COALESCE(updated_by, 0) AS updated_by,
	is_deleted`

func NewTenantRepository(db *pgxpool.Pool) TenantRepository {
	return &PostgresTenantRepository{db: db}
}

func (r *PostgresTenantRepository) GetByID(ctx context.Context, id uint) (*Tenant, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var t Tenant
	err = q.QueryRow(ctx, `
		SELECT`+tenantSelectCols+`
		FROM tenants
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&t.ID, &t.Code, &t.Name, &t.Status,
		&t.Contact, &t.Phone, &t.Email,
		&t.Province, &t.City, &t.Area, &t.Address,
		&t.Config, &t.Dashboard,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CreatedBy, &t.UpdatedBy,
		&t.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTenantRepository) GetByCode(ctx context.Context, code string) (*Tenant, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var t Tenant
	err = q.QueryRow(ctx, `
		SELECT`+tenantSelectCols+`
		FROM tenants
		WHERE is_deleted = FALSE AND code = $1`, code).Scan(
		&t.ID, &t.Code, &t.Name, &t.Status,
		&t.Contact, &t.Phone, &t.Email,
		&t.Province, &t.City, &t.Area, &t.Address,
		&t.Config, &t.Dashboard,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CreatedBy, &t.UpdatedBy,
		&t.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTenantRepository) List(ctx context.Context, keyword string, status *int16, page, size int) ([]Tenant, int64, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, 0, err
	}
	where := "WHERE is_deleted = FALSE"
	args := []interface{}{}
	argIdx := 1

	if keyword != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+keyword+"%")
		argIdx++
	}
	if status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}

	var total int64
	err = q.QueryRow(ctx, "SELECT COUNT(*) FROM tenants "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`SELECT%s FROM tenants %s ORDER BY id ASC LIMIT $%d OFFSET $%d`,
		tenantSelectCols, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(
			&t.ID, &t.Code, &t.Name, &t.Status,
			&t.Contact, &t.Phone, &t.Email,
			&t.Province, &t.City, &t.Area, &t.Address,
			&t.Config, &t.Dashboard,
			&t.CreatedAt, &t.UpdatedAt,
			&t.CreatedBy, &t.UpdatedBy,
			&t.IsDeleted,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, t)
	}
	return list, total, nil
}

func (r *PostgresTenantRepository) Create(ctx context.Context, code, name, contact, phone, email string) (*Tenant, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var t Tenant
	err = q.QueryRow(ctx, `
		INSERT INTO tenants (code, name, contact, phone, email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING`+tenantSelectCols,
		code, name, contact, phone, email,
	).Scan(
		&t.ID, &t.Code, &t.Name, &t.Status,
		&t.Contact, &t.Phone, &t.Email,
		&t.Province, &t.City, &t.Area, &t.Address,
		&t.Config, &t.Dashboard,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CreatedBy, &t.UpdatedBy,
		&t.IsDeleted,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uk_tenants_code") {
			return nil, ErrTenantCodeExists
		}
		return nil, fmt.Errorf("create tenant: %w", err)
	}
	return &t, nil
}

func (r *PostgresTenantRepository) Update(ctx context.Context, id uint, name, contact, phone, email, province, city, area, address string) (*Tenant, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var t Tenant
	err = q.QueryRow(ctx, `
		UPDATE tenants SET
			name = $2, contact = $3, phone = $4, email = $5,
			province = $6, city = $7, area = $8, address = $9,
			updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING`+tenantSelectCols,
		id, name, contact, phone, email,
		province, city, area, address,
	).Scan(
		&t.ID, &t.Code, &t.Name, &t.Status,
		&t.Contact, &t.Phone, &t.Email,
		&t.Province, &t.City, &t.Area, &t.Address,
		&t.Config, &t.Dashboard,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CreatedBy, &t.UpdatedBy,
		&t.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("update tenant: %w", err)
	}
	return &t, nil
}

func (r *PostgresTenantRepository) Delete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `
		UPDATE tenants SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// CountActiveUsers 统计租户下未软删的用户数。
// 注意：必须在 RunInPlatformTx 调用本方法，否则会被 users 表的 RLS 拦截（current_setting('app.bypass_rls') != 'on'）。
func (r *PostgresTenantRepository) CountActiveUsers(ctx context.Context, tenantID uint) (int64, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return 0, err
	}
	var n int64
	err = q.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE tenant_id = $1 AND is_deleted = FALSE`, tenantID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count tenant users: %w", err)
	}
	return n, nil
}

// UpdateStatus 仅修改 status 字段，回填完整行便于审计快照。
func (r *PostgresTenantRepository) UpdateStatus(ctx context.Context, id uint, status int16) (*Tenant, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var t Tenant
	err = q.QueryRow(ctx, `
		UPDATE tenants SET status = $2, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING`+tenantSelectCols,
		id, status,
	).Scan(
		&t.ID, &t.Code, &t.Name, &t.Status,
		&t.Contact, &t.Phone, &t.Email,
		&t.Province, &t.City, &t.Area, &t.Address,
		&t.Config, &t.Dashboard,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CreatedBy, &t.UpdatedBy,
		&t.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("update tenant status: %w", err)
	}
	return &t, nil
}

// purgeOrder 硬删租户数据的顺序。按"先删子表、再删父表"的依赖关系排列，
// 避免 FK 约束冲突。每个表名对应 migrations/framework.sql 中带 tenant_id 的表。
//
// 注意：即使没有显式 FK 约束（migrations 里多数未声明 FK），也按业务依赖排序：
// - usage_records / attachments ：纯记录表，无业务依赖
// - subscriptions：依赖 plans（但 plans 是平台级，不删）
// - role_data_scopes / user_roles / role_menus / role_resources：依赖 roles
// - dict_items：依赖 dicts
// - routes / resources / menus：依赖 organizations / 各 role_resource 关系
// - organizations / roles / users / dicts：核心实体
// - tenant_user_seq：独立序列表
var purgeOrder = []string{
	"usage_records",
	"attachments",
	"subscriptions",
	"role_data_scopes",
	"user_roles",
	"role_menus",
	"role_resources",
	"dict_items",
	"routes",
	"resources",
	"menus",
	"organizations",
	"roles",
	"users",
	"dicts",
	"tenant_user_seq",
}

// PurgeTenantData 硬删所有 tenant_id-bearing 表中的该租户数据。
// 必须在 RunInPlatformTx 内调用，否则会被 RLS 拦截。
// 返回每张表实际删除的行数。
func (r *PostgresTenantRepository) PurgeTenantData(ctx context.Context, tenantID uint) (map[string]int64, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(purgeOrder))
	for _, table := range purgeOrder {
		// 表名来自 migrations 文件常量列表，不是用户输入——安全。
		// 走 $1 参数化防注入（虽然 table 不会被替换）。
		tag, err := q.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = $1`, table), tenantID)
		if err != nil {
			return result, fmt.Errorf("purge %s: %w", table, err)
		}
		result[table] = tag.RowsAffected()
	}
	return result, nil
}

// HardDelete 硬删 tenants 表中指定 id 的行。
// 前置条件：service 层必须先调 PurgeTenantData 清空所有 tenant_id-bearing 表。
func (r *PostgresTenantRepository) HardDelete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("hard delete tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}

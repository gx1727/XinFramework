package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var n int64
	err = q.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_users
		WHERE tenant_id = $1 AND is_deleted = FALSE`, tenantID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count tenant users: %w", err)
	}
	return n, nil
}

// UpdateStatus 仅修改 status 字段，回填完整行便于审计快照。
func (r *PostgresTenantRepository) UpdateStatus(ctx context.Context, id uint, status int16) (*Tenant, error) {
	q, err := db.GetQuerier(ctx, r.db)
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

// GetTenantRecord 把完整 Tenant 适配成 pkg/tenant.TenantRecord（窄接口返回类型）。
// 满足 pkg/tenant.TenantRepository 窄接口（1 method），让 *PostgresTenantRepository
// 不需要 adapter 就能直接写到 AppContext。
func (r *PostgresTenantRepository) GetTenantRecord(ctx context.Context, id uint) (*pkgtenant.TenantRecord, error) {
	t, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pkgtenant.TenantRecord{
		ID:        t.ID,
		Code:      t.Code,
		Name:      t.Name,
		Status:    t.Status,
		Contact:   t.Contact,
		Phone:     t.Phone,
		Email:     t.Email,
		Province:  t.Province,
		City:      t.City,
		Area:      t.Area,
		Address:   t.Address,
		Config:    t.Config,
		Dashboard: t.Dashboard,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}, nil
}

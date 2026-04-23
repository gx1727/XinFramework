package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func scanTenant(row pgx.Row) (*Tenant, error) {
	var t Tenant
	err := row.Scan(
		&t.ID, &t.Code, &t.Name, &t.Status,
		&t.Contact, &t.Phone, &t.Email,
		&t.Province, &t.City, &t.Area, &t.Address,
		&t.Config, &t.Dashboard,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CreatedBy, &t.UpdatedBy,
		&t.IsDeleted,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetByID(ctx context.Context, d *pgxpool.Pool, id uint) (*Tenant, error) {
	return scanTenant(d.QueryRow(ctx, `
		SELECT id, code, name, status, contact, phone, email,
		       province, city, area, address, config, dashboard,
		       created_at, updated_at, created_by, updated_by, is_deleted
		FROM tenants
		WHERE is_deleted = FALSE AND id = $1`, id))
}

func GetByCode(ctx context.Context, d *pgxpool.Pool, code string) (*Tenant, error) {
	return scanTenant(d.QueryRow(ctx, `
		SELECT id, code, name, status, contact, phone, email,
		       province, city, area, address, config, dashboard,
		       created_at, updated_at, created_by, updated_by, is_deleted
		FROM tenants
		WHERE is_deleted = FALSE AND code = $1`, code))
}

func Create(ctx context.Context, d *pgxpool.Pool, req CreateTenantReq) (*Tenant, error) {
	status := int16(1)
	if req.Status != nil {
		status = *req.Status
	}

	var t Tenant
	err := d.QueryRow(ctx, `
		INSERT INTO tenants (code, name, status, contact, phone, email)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, code, name, status, contact, phone, email,
		          province, city, area, address, config, dashboard,
		          created_at, updated_at, created_by, updated_by, is_deleted`,
		req.Code, req.Name, status, req.Contact, req.Phone, req.Email,
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
		return nil, fmt.Errorf("%w: %v", ErrTenantCreateFailed, err)
	}
	return &t, nil
}

func Update(ctx context.Context, d *pgxpool.Pool, id uint, req UpdateTenantReq) (*Tenant, error) {
	status := int16(1)
	if req.Status != nil {
		status = *req.Status
	}

	t, err := scanTenant(d.QueryRow(ctx, `
		UPDATE tenants SET
			name = $2, status = $3, contact = $4, phone = $5, email = $6,
			province = $7, city = $8, area = $9, address = $10,
			updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, code, name, status, contact, phone, email,
		          province, city, area, address, config, dashboard,
		          created_at, updated_at, created_by, updated_by, is_deleted`,
		id, req.Name, status, req.Contact, req.Phone, req.Email,
		req.Province, req.City, req.Area, req.Address,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrTenantUpdateFailed, err)
	}
	return t, nil
}

func Delete(ctx context.Context, d *pgxpool.Pool, id uint) error {
	tag, err := d.Exec(ctx, `
		UPDATE tenants SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTenantDeleteFailed, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}

func List(ctx context.Context, d *pgxpool.Pool, req ListTenantReq) ([]Tenant, int64, error) {
	where := "WHERE is_deleted = FALSE"
	args := []interface{}{}
	argIdx := 1

	if req.Keyword != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+req.Keyword+"%")
		argIdx++
	}
	if req.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *req.Status)
		argIdx++
	}

	var total int64
	err := d.QueryRow(ctx, "SELECT COUNT(*) FROM tenants "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, ErrTenantListFailed
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`
		SELECT id, code, name, status, contact, phone, email,
		       province, city, area, address, config, dashboard,
		       created_at, updated_at, created_by, updated_by, is_deleted
		FROM tenants %s
		ORDER BY id ASC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := d.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, ErrTenantListFailed
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
			return nil, 0, ErrTenantListFailed
		}
		list = append(list, t)
	}
	return list, total, nil
}

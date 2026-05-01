package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTenantRepository implements TenantRepository
type PostgresTenantRepository struct {
	db *pgxpool.Pool
}

func NewTenantRepository(db *pgxpool.Pool) TenantRepository {
	return &PostgresTenantRepository{db: db}
}

func (r *PostgresTenantRepository) GetByID(ctx context.Context, id uint) (*Tenant, error) {
	var t Tenant
	err := r.db.QueryRow(ctx, `
		SELECT id, code, name, status, contact, phone, email,
		       province, city, area, address, config, dashboard,
		       created_at, updated_at, created_by, updated_by, is_deleted
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
	var t Tenant
	err := r.db.QueryRow(ctx, `
		SELECT id, code, name, status, contact, phone, email,
		       province, city, area, address, config, dashboard,
		       created_at, updated_at, created_by, updated_by, is_deleted
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
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM tenants "+where, args...).Scan(&total)
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

	query := fmt.Sprintf(`SELECT id, code, name, status, contact, phone, email,
	       province, city, area, address, config, dashboard,
	       created_at, updated_at, created_by, updated_by, is_deleted
		FROM tenants %s ORDER BY id ASC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.Query(ctx, query, args...)
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
	var t Tenant
	err := r.db.QueryRow(ctx, `
		INSERT INTO tenants (code, name, contact, phone, email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, name, status, contact, phone, email,
		          province, city, area, address, config, dashboard,
		          created_at, updated_at, created_by, updated_by, is_deleted`,
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
	var t Tenant
	err := r.db.QueryRow(ctx, `
		UPDATE tenants SET
			name = $2, contact = $3, phone = $4, email = $5,
			province = $6, city = $7, area = $8, address = $9,
			updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, code, name, status, contact, phone, email,
		          province, city, area, address, config, dashboard,
		          created_at, updated_at, created_by, updated_by, is_deleted`,
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
	tag, err := r.db.Exec(ctx, `
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

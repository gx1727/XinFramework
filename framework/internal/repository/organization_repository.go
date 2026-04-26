package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/model"
)

type PostgresOrganizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) model.OrganizationRepository {
	return &PostgresOrganizationRepository{db: db}
}

func (r *PostgresOrganizationRepository) GetByID(ctx context.Context, id uint) (*model.Organization, error) {
	var org model.Organization
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM organizations
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
		&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}
	return &org, nil
}

func (r *PostgresOrganizationRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*model.Organization, error) {
	var org model.Organization
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM organizations
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
		&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}
	return &org, nil
}

func (r *PostgresOrganizationRepository) GetByTenant(ctx context.Context, tenantID uint) ([]model.Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM organizations
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var org model.Organization
		if err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
			&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (r *PostgresOrganizationRepository) GetChildren(ctx context.Context, parentID uint) ([]model.Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM organizations
		WHERE is_deleted = FALSE AND parent_id = $1
		ORDER BY sort ASC, id ASC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var org model.Organization
		if err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
			&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (r *PostgresOrganizationRepository) GetTree(ctx context.Context, tenantID uint) ([]model.Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM organizations
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY ancestors ASC, sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var org model.Organization
		if err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
			&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (r *PostgresOrganizationRepository) Create(ctx context.Context, tenantID uint, req model.CreateOrgRepoReq) (*model.Organization, error) {
	var org model.Organization
	err := r.db.QueryRow(ctx, `
		INSERT INTO organizations (tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
	`, tenantID, req.Code, req.Name, req.Type, req.Description, req.AdminCode, req.ParentID, req.Ancestors, req.Sort, req.Status).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
		&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	return &org, nil
}

func (r *PostgresOrganizationRepository) Update(ctx context.Context, id uint, req model.UpdateOrgRepoReq) (*model.Organization, error) {
	var org model.Organization
	err := r.db.QueryRow(ctx, `
		UPDATE organizations SET name = $2, type = $3, description = $4, admin_code = $5, sort = $6, status = $7, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
	`, id, req.Name, req.Type, req.Description, req.AdminCode, req.Sort, req.Status).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.Type, &org.Description,
		&org.AdminCode, &org.ParentID, &org.Ancestors, &org.Sort, &org.Status, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, fmt.Errorf("update organization: %w", err)
	}
	return &org, nil
}

func (r *PostgresOrganizationRepository) Delete(ctx context.Context, id uint) error {
	_, err := r.db.Exec(ctx, `UPDATE organizations SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

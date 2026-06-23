package organization

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/permission"
)

type PostgresOrganizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) OrganizationRepository {
	return &PostgresOrganizationRepository{db: db}
}

var organizationScopeColumns = permission.ScopeColumns{
	SelfColumn:   "id",
	SelfUseOrgID: true,
	OrgID:        "id",
}

func buildOrganizationScopeFilter(ctx context.Context) (permission.ScopeFilter, error) {
	uc, ok := xincontext.UserContextFrom(ctx)
	if !ok || uc == nil || uc.UserID == 0 {
		return permission.ScopeFilter{}, nil
	}
	return uc.GetDataScopeFilterFor(organizationScopeColumns)
}

func rebindScopeSQL(sql string, from, to int) string {
	for i := from; i >= 1; i-- {
		sql = strings.ReplaceAll(sql, fmt.Sprintf("$%d", i), fmt.Sprintf("$%d", to+i-1))
	}
	return sql
}

func (r *PostgresOrganizationRepository) GetByID(ctx context.Context, id uint) (*Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var org Organization
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
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

func (r *PostgresOrganizationRepository) GetByIDScoped(ctx context.Context, id uint) (*Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	filter, err := buildOrganizationScopeFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND id = $1`
	args := []any{id}
	if !filter.IsEmpty() {
		query += " AND (" + rebindScopeSQL(filter.SQL, len(filter.Args), 2) + ")"
		args = append(args, filter.Args...)
	}

	var org Organization
	err = q.QueryRow(ctx, query, args...).Scan(
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

func (r *PostgresOrganizationRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var org Organization
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
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

func (r *PostgresOrganizationRepository) GetByTenant(ctx context.Context, tenantID uint) ([]Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
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

func (r *PostgresOrganizationRepository) GetByTenantScoped(ctx context.Context, tenantID uint) ([]Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	filter, err := buildOrganizationScopeFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND tenant_id = $1`
	args := []any{tenantID}
	if !filter.IsEmpty() {
		query += " AND (" + rebindScopeSQL(filter.SQL, len(filter.Args), 2) + ")"
		args = append(args, filter.Args...)
	}
	query += " ORDER BY sort ASC, id ASC"

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
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

// CountChildren 统计 parentID 下未删子组织数（不含自己）。
// CountUsersInOrgTree 统计本组织及其所有后代下的未删除用户数。
// 用 ancestors 字符串前缀匹配才能一次扫到后代，不依赖 ltree 扩展。
func (r *PostgresOrganizationRepository) CountUsersInOrgTree(ctx context.Context, orgID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return 0, err
	}

	var n int64
	// o.ancestors 存的是"父轨迹"，如 "0.1.2"；orgID 本身需额外匹配。
	err = q.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM tenant_users u
		JOIN tenant_organizations o ON o.id = u.org_id AND o.is_deleted = FALSE
		WHERE u.is_deleted = FALSE
		  AND (o.id = $1 OR o.ancestors LIKE $2)`,
		orgID, fmt.Sprintf("%d.", orgID)).Scan(&n)
	return n, err
}

func (r *PostgresOrganizationRepository) CountChildren(ctx context.Context, parentID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var n int64
	err = q.QueryRow(ctx, `SELECT COUNT(*) FROM tenant_organizations WHERE is_deleted = FALSE AND parent_id = $1`, parentID).Scan(&n)
	return n, err
}

func (r *PostgresOrganizationRepository) GetChildren(ctx context.Context, parentID uint) ([]Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND parent_id = $1
		ORDER BY sort ASC, id ASC`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
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

func (r *PostgresOrganizationRepository) GetChildrenScoped(ctx context.Context, parentID uint) ([]Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	filter, err := buildOrganizationScopeFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND parent_id = $1`
	args := []any{parentID}
	if !filter.IsEmpty() {
		query += " AND (" + rebindScopeSQL(filter.SQL, len(filter.Args), 2) + ")"
		args = append(args, filter.Args...)
	}
	query += " ORDER BY sort ASC, id ASC"

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
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

func (r *PostgresOrganizationRepository) GetTree(ctx context.Context, tenantID uint) ([]Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY ancestors ASC, sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
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

func (r *PostgresOrganizationRepository) GetTreeScoped(ctx context.Context, tenantID uint) ([]Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	filter, err := buildOrganizationScopeFilter(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status, created_at, updated_at
		FROM tenant_organizations
		WHERE is_deleted = FALSE AND tenant_id = $1`
	args := []any{tenantID}
	if !filter.IsEmpty() {
		query += " AND (" + rebindScopeSQL(filter.SQL, len(filter.Args), 2) + ")"
		args = append(args, filter.Args...)
	}
	query += " ORDER BY ancestors ASC, sort ASC, id ASC"

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
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

func (r *PostgresOrganizationRepository) Create(ctx context.Context, tenantID uint, req CreateOrgRepoReq) (*Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var org Organization
	err = q.QueryRow(ctx, `
		INSERT INTO tenant_organizations (tenant_id, code, name, type, description, admin_code, parent_id, ancestors, sort, status)
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

func (r *PostgresOrganizationRepository) Update(ctx context.Context, id uint, req UpdateOrgRepoReq) (*Organization, error) {
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var org Organization
	err = q.QueryRow(ctx, `
		UPDATE tenant_organizations SET name = $2, type = $3, description = $4, admin_code = $5, sort = $6, status = $7, updated_at = NOW()
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
	q, err := db.GetQuerier(ctx, r.db)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE tenant_organizations SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrOrgNotFound
	}
	return nil
}

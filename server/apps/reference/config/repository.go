// Package config 通用配置 - PostgreSQL 仓储
package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

type PostgresConfigRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresConfigRepository(pool *pgxpool.Pool) *PostgresConfigRepository {
	return &PostgresConfigRepository{pool: pool}
}

// =============== Group ===============

func (r *PostgresConfigRepository) ListGroups(ctx context.Context, tenantID uint) ([]ConfigGroup, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, description, icon, sort, is_system, is_public, status, created_at, updated_at
		FROM config_groups
		WHERE is_deleted = FALSE AND tenant_id IN (0, $1)
		ORDER BY sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()
	list := make([]ConfigGroup, 0)
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *PostgresConfigRepository) GetGroupByID(ctx context.Context, id uint) (*ConfigGroup, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, description, icon, sort, is_system, is_public, status, created_at, updated_at
		FROM config_groups WHERE is_deleted = FALSE AND id = $1`, id)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresConfigRepository) GetGroupByCode(ctx context.Context, tenantID uint, code string) (*ConfigGroup, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, description, icon, sort, is_system, is_public, status, created_at, updated_at
		FROM config_groups
		WHERE is_deleted = FALSE AND tenant_id IN (0, $1) AND code = $2
		ORDER BY tenant_id DESC LIMIT 1`, tenantID, code)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresConfigRepository) CreateGroup(ctx context.Context, tenantID uint, req CreateGroupRepoReq) (*ConfigGroup, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		INSERT INTO config_groups (tenant_id, code, name, description, icon, sort, is_system, is_public, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1)
		RETURNING id, tenant_id, code, name, description, icon, sort, is_system, is_public, status, created_at, updated_at`,
		tenantID, req.Code, req.Name, req.Description, req.Icon, req.Sort, req.IsSystem, req.IsPublic)
	g, err := scanGroup(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrGroupCodeExists
		}
		return nil, fmt.Errorf("create group: %w", err)
	}
	return &g, nil
}

func (r *PostgresConfigRepository) UpdateGroup(ctx context.Context, id uint, req UpdateGroupRepoReq) (*ConfigGroup, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		UPDATE config_groups SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			icon        = COALESCE($4, icon),
			sort        = COALESCE($5, sort),
			is_public   = COALESCE($6, is_public),
			status      = COALESCE($7, status),
			updated_at  = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, code, name, description, icon, sort, is_system, is_public, status, created_at, updated_at`,
		id, req.Name, req.Description, req.Icon, req.Sort, req.IsPublic, req.Status)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("update group: %w", err)
	}
	return &g, nil
}

func (r *PostgresConfigRepository) DeleteGroup(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE config_groups SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrGroupNotFound
	}
	return nil
}

// =============== Item ===============

func (r *PostgresConfigRepository) ListItemsByGroup(ctx context.Context, groupID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, group_id, key, value, default_value, type, label, description,
		       options, validation, sort, is_public, is_readonly, is_system, status, created_at, updated_at
		FROM config_items
		WHERE is_deleted = FALSE AND group_id = $1
		ORDER BY sort ASC, id ASC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()
	list := make([]ConfigItem, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, nil
}

func (r *PostgresConfigRepository) ListItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, group_id, key, value, default_value, type, label, description,
		       options, validation, sort, is_public, is_readonly, is_system, status, created_at, updated_at
		FROM config_items
		WHERE is_deleted = FALSE AND tenant_id IN (0, $1)
		ORDER BY group_id ASC, sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list tenant items: %w", err)
	}
	defer rows.Close()
	list := make([]ConfigItem, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, nil
}

func (r *PostgresConfigRepository) ListPublicItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT ci.id, ci.tenant_id, ci.group_id, ci.key, ci.value, ci.default_value, ci.type, ci.label, ci.description,
		       ci.options, ci.validation, ci.sort, ci.is_public, ci.is_readonly, ci.is_system, ci.status, ci.created_at, ci.updated_at
		FROM config_items ci
		JOIN config_groups cg ON cg.id = ci.group_id AND cg.is_deleted = FALSE
		WHERE ci.is_deleted = FALSE AND ci.tenant_id IN (0, $1)
		  AND ci.is_public = TRUE AND cg.is_public = TRUE
		ORDER BY cg.sort ASC, ci.sort ASC, ci.id ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list public items: %w", err)
	}
	defer rows.Close()
	list := make([]ConfigItem, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, nil
}

func (r *PostgresConfigRepository) ListPublicItemsByGroupCode(ctx context.Context, tenantID uint, groupCode string) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT ci.id, ci.tenant_id, ci.group_id, ci.key, ci.value, ci.default_value, ci.type, ci.label, ci.description,
		       ci.options, ci.validation, ci.sort, ci.is_public, ci.is_readonly, ci.is_system, ci.status, ci.created_at, ci.updated_at
		FROM config_items ci
		JOIN config_groups cg ON cg.id = ci.group_id AND cg.is_deleted = FALSE
		WHERE ci.is_deleted = FALSE AND ci.tenant_id IN (0, $1)
		  AND ci.is_public = TRUE AND cg.is_public = TRUE
		  AND cg.code = $2
		ORDER BY ci.sort ASC, ci.id ASC`, tenantID, groupCode)
	if err != nil {
		return nil, fmt.Errorf("list public items by group: %w", err)
	}
	defer rows.Close()
	list := make([]ConfigItem, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, nil
}

func (r *PostgresConfigRepository) GetItemByID(ctx context.Context, id uint) (*ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, group_id, key, value, default_value, type, label, description,
		       options, validation, sort, is_public, is_readonly, is_system, status, created_at, updated_at
		FROM config_items WHERE is_deleted = FALSE AND id = $1`, id)
	it, err := scanItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *PostgresConfigRepository) CreateItem(ctx context.Context, tenantID, groupID uint, req CreateItemRepoReq) (*ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	valueJSON := toJSON(req.Value)
	defaultValueJSON := toJSON(req.DefaultValue)
	optionsJSON := toJSON(req.Options)
	validationJSON := toJSON(req.Validation)
	row := q.QueryRow(ctx, `
		INSERT INTO config_items
		    (tenant_id, group_id, key, value, default_value, type, label, description, options, validation,
		     sort, is_public, is_readonly, is_system, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 1)
		RETURNING id, tenant_id, group_id, key, value, default_value, type, label, description,
		          options, validation, sort, is_public, is_readonly, is_system, status, created_at, updated_at`,
		tenantID, groupID, req.Key, valueJSON, defaultValueJSON, req.Type, req.Label, req.Description,
		optionsJSON, validationJSON, req.Sort, req.IsPublic, req.IsReadonly, req.IsSystem)
	it, err := scanItem(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrItemKeyExists
		}
		return nil, fmt.Errorf("create item: %w", err)
	}
	return &it, nil
}

func (r *PostgresConfigRepository) UpdateItem(ctx context.Context, id uint, req UpdateItemRepoReq) (*ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	var valueJSON any
	if req.Value != nil {
		valueJSON = toJSON(*req.Value)
	}
	row := q.QueryRow(ctx, `
		UPDATE config_items SET
			value       = COALESCE($2, value),
			label       = COALESCE($3, label),
			description = COALESCE($4, description),
			sort        = COALESCE($5, sort),
			is_public   = COALESCE($6, is_public),
			is_readonly = COALESCE($7, is_readonly),
			status      = COALESCE($8, status),
			updated_at  = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, group_id, key, value, default_value, type, label, description,
		          options, validation, sort, is_public, is_readonly, is_system, status, created_at, updated_at`,
		id, valueJSON, req.Label, req.Description, req.Sort, req.IsPublic, req.IsReadonly, req.Status)
	it, err := scanItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("update item: %w", err)
	}
	return &it, nil
}

func (r *PostgresConfigRepository) ResetItem(ctx context.Context, id uint) (*ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		UPDATE config_items SET
			value      = default_value,
			updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, group_id, key, value, default_value, type, label, description,
		          options, validation, sort, is_public, is_readonly, is_system, status, created_at, updated_at`, id)
	it, err := scanItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("reset item: %w", err)
	}
	return &it, nil
}

func (r *PostgresConfigRepository) DeleteItem(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE config_items SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *PostgresConfigRepository) CountItemsByGroup(ctx context.Context, groupID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return 0, err
	}
	var n int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM config_items WHERE is_deleted = FALSE AND group_id = $1`, groupID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return n, nil
}

// =============== helpers ===============

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGroup(row rowScanner) (ConfigGroup, error) {
	var g ConfigGroup
	if err := row.Scan(
		&g.ID, &g.TenantID, &g.Code, &g.Name, &g.Description, &g.Icon, &g.Sort,
		&g.IsSystem, &g.IsPublic, &g.Status, &g.CreatedAt, &g.UpdatedAt,
	); err != nil {
		return ConfigGroup{}, err
	}
	return g, nil
}

func scanItem(row rowScanner) (ConfigItem, error) {
	var it ConfigItem
	var valueJSON, defaultValueJSON, optionsJSON, validationJSON []byte
	if err := row.Scan(
		&it.ID, &it.TenantID, &it.GroupID, &it.Key, &valueJSON, &defaultValueJSON,
		&it.Type, &it.Label, &it.Description, &optionsJSON, &validationJSON,
		&it.Sort, &it.IsPublic, &it.IsReadonly, &it.IsSystem, &it.Status,
		&it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		return ConfigItem{}, err
	}
	if len(valueJSON) > 0 {
		_ = json.Unmarshal(valueJSON, &it.Value)
	}
	if len(defaultValueJSON) > 0 {
		_ = json.Unmarshal(defaultValueJSON, &it.DefaultValue)
	}
	if len(optionsJSON) > 0 {
		_ = json.Unmarshal(optionsJSON, &it.Options)
	}
	if len(validationJSON) > 0 {
		_ = json.Unmarshal(validationJSON, &it.Validation)
	}
	return it, nil
}

func toJSON(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// Package config 通用配置 - PostgreSQL 仓储
//
//   - SELECT 包含 scope / visibility / extend
//   - CreateGroup 加 scope 参数
//   - 加 platform / tenant 区分的查询方法
//   - 加 override CRUD
//   - 加 visibility CRUD
//   - 加 Resolve 方法（业务合并消费）
package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

const groupSelectCols = `
	id, tenant_id, code, name, description, icon, sort,
	scope, visibility,
	is_system, is_public, status,
	COALESCE(extend, '{}'::jsonb)::text as extend,
	created_at, updated_at`

const itemSelectCols = `
	id, tenant_id, category_id, key, value, default_value,
	type, label, description, options, validation,
	sort, is_public, is_readonly, is_system,
	platform_item_id, is_override,
	status, created_at, updated_at`

// itemSelectColsPrefixed 与 itemSelectCols 同序，但每列带 `ci.` 别名。
// 专用于同时 JOIN config_categories cg 的查询（避免与 cg.id / cg.tenant_id /
// cg.description / cg.sort / cg.status 等同名列冲突，PG 会报 SQLSTATE 42702）。
// 与单表查询里使用的 itemSelectCols 不可互换。
const itemSelectColsPrefixed = `
	ci.id, ci.tenant_id, ci.category_id, ci.key, ci.value, ci.default_value,
	ci.type, ci.label, ci.description, ci.options, ci.validation,
	ci.sort, ci.is_public, ci.is_readonly, ci.is_system,
	ci.platform_item_id, ci.is_override,
	ci.status, ci.created_at, ci.updated_at`

func scanGroup(row pgx.Row) (ConfigCategory, error) {
	var g ConfigCategory
	var extendStr string
	err := row.Scan(
		&g.ID, &g.TenantID, &g.Code, &g.Name,
		&g.Description, &g.Icon, &g.Sort,
		&g.Scope, &g.Visibility,
		&g.IsSystem, &g.IsPublic, &g.Status,
		&extendStr,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return g, err
	}
	if extendStr != "" {
		_ = json.Unmarshal([]byte(extendStr), &g.Extend)
	}
	return g, nil
}

func scanItem(row pgx.Row) (ConfigItem, error) {
	var item ConfigItem
	var valueJSON, defaultValueJSON, optionsJSON, validationJSON []byte
	var label, desc *string
	err := row.Scan(
		&item.ID, &item.TenantID, &item.CategoryID, &item.Key,
		&valueJSON, &defaultValueJSON,
		&item.Type, &label, &desc, &optionsJSON, &validationJSON,
		&item.Sort, &item.IsPublic, &item.IsReadonly, &item.IsSystem,
		&item.PlatformItemID, &item.IsOverride,
		&item.Status, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return item, err
	}
	item.Label = label
	item.Description = desc
	if len(valueJSON) > 0 {
		_ = json.Unmarshal(valueJSON, &item.Value)
	}
	if len(defaultValueJSON) > 0 {
		_ = json.Unmarshal(defaultValueJSON, &item.DefaultValue)
	}
	if len(optionsJSON) > 0 {
		_ = json.Unmarshal(optionsJSON, &item.Options)
	}
	if len(validationJSON) > 0 {
		_ = json.Unmarshal(validationJSON, &item.Validation)
	}
	return item, nil
}

// ============================================================================
// Group — Tenant 域
// ============================================================================

func (r *PostgresConfigRepository) ListGroups(ctx context.Context, tenantID uint) ([]ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()
	list := make([]ConfigCategory, 0)
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *PostgresConfigRepository) GetGroupByID(ctx context.Context, id uint) (*ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories WHERE is_deleted = FALSE AND id = $1`, id)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresConfigRepository) GetGroupByCode(ctx context.Context, tenantID uint, code string) (*ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories
		WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2
		LIMIT 1`, tenantID, code)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresConfigRepository) CreateGroup(ctx context.Context, tenantID uint, scope string, req CreateGroupRepoReq) (*ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(map[string]any{})
	row := q.QueryRow(ctx, `
		INSERT INTO config_categories
			(tenant_id, code, name, description, icon, sort, scope, visibility, is_system, is_public, status, extend)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'all', $8, $9, 1, $10)
		RETURNING `+groupSelectCols,
		tenantID, req.Code, req.Name, req.Description, req.Icon, req.Sort,
		scope, req.IsSystem, req.IsPublic, extendJSON)
	g, err := scanGroup(row)
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, ErrGroupCodeExists
		}
		return nil, fmt.Errorf("create group: %w", err)
	}
	return &g, nil
}

func (r *PostgresConfigRepository) UpdateGroup(ctx context.Context, id uint, req UpdateGroupRepoReq) (*ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		UPDATE config_categories SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			icon        = COALESCE($4, icon),
			sort        = COALESCE($5, sort),
			visibility  = COALESCE($6, visibility),
			is_public   = COALESCE($7, is_public),
			status      = COALESCE($8, status),
			updated_at  = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING `+groupSelectCols,
		id, req.Name, req.Description, req.Icon, req.Sort, req.Visibility, req.IsPublic, req.Status)
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
	tag, err := q.Exec(ctx, `
		UPDATE config_categories SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrGroupNotFound
	}
	return nil
}

// ============================================================================
// Group — Platform 域
// ============================================================================

func (r *PostgresConfigRepository) ListPlatformGroups(ctx context.Context) ([]ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories
		WHERE is_deleted = FALSE AND scope = 'platform'
		ORDER BY sort ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]ConfigCategory, 0)
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *PostgresConfigRepository) GetPlatformGroupByCode(ctx context.Context, code string) (*ConfigCategory, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories
		WHERE is_deleted = FALSE AND scope = 'platform' AND code = $1
		LIMIT 1`, code)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &g, nil
}

// ============================================================================
// Item
// ============================================================================

func (r *PostgresConfigRepository) ListItemsByGroup(ctx context.Context, categoryID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+itemSelectCols+`
		FROM config_items
		WHERE is_deleted = FALSE AND category_id = $1
		ORDER BY sort ASC, id ASC`, categoryID)
	if err != nil {
		return nil, err
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
		SELECT `+itemSelectCols+`
		FROM config_items
		WHERE is_deleted = FALSE AND tenant_id = $1
		ORDER BY category_id, sort ASC, id ASC`, tenantID)
	if err != nil {
		return nil, err
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

func (r *PostgresConfigRepository) ListPlatformItemsByGroup(ctx context.Context, categoryID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+itemSelectCols+`
		FROM config_items
		WHERE is_deleted = FALSE AND tenant_id = 0 AND category_id = $1
		ORDER BY sort ASC, id ASC`, categoryID)
	if err != nil {
		return nil, err
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
	// JOIN config_categories cg → 必须用 itemSelectColsPrefixed（带 ci.）避免 SQLSTATE 42702
	rows, err := q.Query(ctx, `
		SELECT `+itemSelectColsPrefixed+`
		FROM config_items ci
		JOIN config_categories cg ON cg.id = ci.category_id
		WHERE ci.is_deleted = FALSE AND cg.is_deleted = FALSE AND cg.is_public = TRUE
		  AND ci.tenant_id = $1 AND ci.is_public = TRUE
		ORDER BY ci.category_id, ci.sort ASC, ci.id ASC`, tenantID)
	if err != nil {
		return nil, err
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
	// JOIN config_categories cg → 必须用 itemSelectColsPrefixed（带 ci.）避免 SQLSTATE 42702
	rows, err := q.Query(ctx, `
		SELECT `+itemSelectColsPrefixed+`
		FROM config_items ci
		JOIN config_categories cg ON cg.id = ci.category_id
		WHERE ci.is_deleted = FALSE AND cg.is_deleted = FALSE AND cg.is_public = TRUE
		  AND ci.tenant_id = $1 AND ci.is_public = TRUE AND cg.code = $2
		ORDER BY ci.sort ASC, ci.id ASC`, tenantID, groupCode)
	if err != nil {
		return nil, err
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
		SELECT `+itemSelectCols+`
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

func (r *PostgresConfigRepository) CreateItem(ctx context.Context, tenantID, categoryID uint, req CreateItemRepoReq) (*ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	valueJSON, _ := json.Marshal(req.Value)
	defaultValueJSON, _ := json.Marshal(req.DefaultValue)
	optionsJSON, _ := json.Marshal(req.Options)
	validationJSON, _ := json.Marshal(req.Validation)

	row := q.QueryRow(ctx, `
		INSERT INTO config_items
			(tenant_id, category_id, key, value, default_value, type, label, description, options, validation,
			 sort, is_public, is_readonly, is_system, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		        $11, $12, $13, $14, 1)
		RETURNING `+itemSelectCols,
		tenantID, categoryID, req.Key, valueJSON, defaultValueJSON, req.Type, req.Label, req.Description,
		optionsJSON, validationJSON,
		req.Sort, req.IsPublic, req.IsReadonly, req.IsSystem)
	it, err := scanItem(row)
	if err != nil {
		if isPgUniqueViolation(err) {
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
	// Value 字段特殊处理：JSONB 显式 cast
	valueJSON, _ := json.Marshal(req.Value)

	row := q.QueryRow(ctx, `
		UPDATE config_items SET
			value       = COALESCE($2::jsonb, value),
			label       = COALESCE($3, label),
			description = COALESCE($4, description),
			sort        = COALESCE($5, sort),
			is_public   = COALESCE($6, is_public),
			is_readonly = COALESCE($7, is_readonly),
			status      = COALESCE($8, status),
			updated_at  = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING `+itemSelectCols,
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
		UPDATE config_items SET value = default_value, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING `+itemSelectCols, id)
	it, err := scanItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *PostgresConfigRepository) DeleteItem(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `
		UPDATE config_items SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *PostgresConfigRepository) CountItemsByGroup(ctx context.Context, categoryID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return 0, err
	}
	var n int64
	err = q.QueryRow(ctx, `
		SELECT COUNT(*) FROM config_items
		WHERE is_deleted = FALSE AND category_id = $1`, categoryID).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ============================================================================
// Override（租户覆盖 platform item）
// ============================================================================

func (r *PostgresConfigRepository) UpsertOverride(ctx context.Context, tenantID, platformItemID uint, value any) (*ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	valueJSON, _ := json.Marshal(value)
	row := q.QueryRow(ctx, `
		INSERT INTO config_items
			(tenant_id, category_id, key, value, type, platform_item_id, is_override, status)
		SELECT $1, ci.category_id, ci.key, $2::jsonb, ci.type, ci.id, TRUE, 1
		FROM config_items ci
		WHERE ci.id = $3 AND ci.tenant_id = 0 AND ci.is_deleted = FALSE
		ON CONFLICT (tenant_id, platform_item_id) WHERE is_override = TRUE AND is_deleted = FALSE
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		RETURNING `+itemSelectCols, tenantID, valueJSON, platformItemID)
	it, err := scanItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlatformItemMismatch
		}
		return nil, err
	}
	return &it, nil
}

func (r *PostgresConfigRepository) DeleteOverride(ctx context.Context, tenantID, platformItemID uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `
		UPDATE config_items SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND is_override = TRUE
		  AND tenant_id = $1 AND platform_item_id = $2`, tenantID, platformItemID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

// ============================================================================
// Visibility（平台 group 对租户的访问级别）
// ============================================================================

func (r *PostgresConfigRepository) ListVisibility(ctx context.Context, categoryID uint) ([]ConfigVisibility, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, category_id, tenant_id, access, created_at, updated_at
		FROM config_visibility WHERE category_id = $1`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]ConfigVisibility, 0)
	for rows.Next() {
		var v ConfigVisibility
		if err := rows.Scan(&v.ID, &v.CategoryID, &v.TenantID, &v.Access, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *PostgresConfigRepository) UpsertVisibility(ctx context.Context, categoryID, tenantID uint, access string) (*ConfigVisibility, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		INSERT INTO config_visibility (category_id, tenant_id, access)
		VALUES ($1, $2, $3)
		ON CONFLICT (category_id, tenant_id) DO UPDATE SET access = EXCLUDED.access, updated_at = NOW()
		RETURNING id, category_id, tenant_id, access, created_at, updated_at`,
		categoryID, tenantID, access)
	var v ConfigVisibility
	if err := row.Scan(&v.ID, &v.CategoryID, &v.TenantID, &v.Access, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *PostgresConfigRepository) DeleteVisibility(ctx context.Context, categoryID, tenantID uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `DELETE FROM config_visibility WHERE category_id = $1 AND tenant_id = $2`, categoryID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrGroupNotFound
	}
	return nil
}

// ============================================================================
// Resolve（业务合并消费）
// ============================================================================

// ResolveGroupForTenant 返回租户视角的合并配置
//
// 合并逻辑：
//  1. 找 platform scope 的 group（code 匹配）
//  2. 检查 visibility：access='invisible' 直接返回 ErrGroupInvisible
//  3. 检查 config_visibility：access 限制
//  4. 拉 platform items
//  5. 拉 tenant override items（is_override=TRUE）
//  6. 合并：override 覆盖 platform
func (r *PostgresConfigRepository) ResolveGroupForTenant(ctx context.Context, tenantID uint, groupCode string) (*ResolvedConfig, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}

	// 1) 找 platform group
	var group ConfigCategory
	err = q.QueryRow(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories
		WHERE is_deleted = FALSE AND scope = 'platform' AND code = $1
		LIMIT 1`, groupCode).Scan( /* ... */ )
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	// 注：scanGroup 用 Scan，但 QueryRow().Scan() 需要逐字段 Scan。
	// 为简化这里复用 scanGroup 对 pgx.Row 的支持（scanGroup 接收 pgx.Row）。
	// 重新写一次：
	group, err = scanGroupFromQueryRow(q.QueryRow(ctx, `
		SELECT `+groupSelectCols+`
		FROM config_categories
		WHERE is_deleted = FALSE AND scope = 'platform' AND code = $1
		LIMIT 1`, groupCode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	// 2) 检查 visibility（简化版：仅 visibility='all' 全开）
	if group.Visibility == "whitelist" {
		// 简化：实际应该查 config_visibility 决定哪些 tenant 在白名单
		// MVP 不实现严格的白名单/黑名单
	}

	// 3) 拉 platform items
	platformItems, err := r.listPlatformItemsByGroupInternal(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	// 4) 拉 tenant override items
	overrideItems, err := r.listOverrideItemsInternal(ctx, tenantID, group.ID)
	if err != nil {
		return nil, err
	}

	// 5) 合并
	out := &ResolvedConfig{
		CategoryID:   group.ID,
		CategoryCode: group.Code,
		CategoryName: group.Name,
		Items:        make(map[string]ResolvedItem, len(platformItems)),
	}
	for _, it := range platformItems {
		out.Items[it.Key] = ResolvedItem{
			Key: it.Key, Value: it.Value, Type: it.Type, Label: it.Label,
			PlatformItemID: nil, IsOverride: false, Source: "platform",
		}
	}
	for _, it := range overrideItems {
		platformID := it.PlatformItemID
		out.Items[it.Key] = ResolvedItem{
			Key: it.Key, Value: it.Value, Type: it.Type, Label: it.Label,
			PlatformItemID: platformID, IsOverride: true, Source: "override",
		}
	}
	return out, nil
}

// ResolveAllForTenant 返回租户全部合并配置（map[group_code]*ResolvedConfig）
func (r *PostgresConfigRepository) ResolveAllForTenant(ctx context.Context, tenantID uint) (map[string]*ResolvedConfig, error) {
	platformGroups, err := r.ListPlatformGroups(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]*ResolvedConfig, len(platformGroups))
	for _, g := range platformGroups {
		rc, err := r.ResolveGroupForTenant(ctx, tenantID, g.Code)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				continue
			}
			return nil, err
		}
		out[g.Code] = rc
	}
	return out, nil
}

// 内部辅助方法（避免重复 Query 构造）
func (r *PostgresConfigRepository) listPlatformItemsByGroupInternal(ctx context.Context, categoryID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+itemSelectCols+`
		FROM config_items
		WHERE is_deleted = FALSE AND tenant_id = 0 AND category_id = $1
		ORDER BY sort ASC, id ASC`, categoryID)
	if err != nil {
		return nil, err
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

func (r *PostgresConfigRepository) listOverrideItemsInternal(ctx context.Context, tenantID, categoryID uint) ([]ConfigItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT `+itemSelectCols+`
		FROM config_items
		WHERE is_deleted = FALSE AND is_override = TRUE
		  AND tenant_id = $1 AND category_id = $2
		ORDER BY sort ASC, id ASC`, tenantID, categoryID)
	if err != nil {
		return nil, err
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

// scanGroupFromQueryRow 是 scanGroup 的 pgx.Row 别名。
// pgx.Row 接口和 pgx.Rows 接口的 Scan 方法签名相同，
// 所以可以直接复用 scanGroup。
func scanGroupFromQueryRow(row pgx.Row) (ConfigCategory, error) {
	return scanGroup(row)
}

// isPgUniqueViolation 判断是否 PG unique_violation（23505）
func isPgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// strings.Contains 引用（防止 unused import）
var _ = strings.Contains

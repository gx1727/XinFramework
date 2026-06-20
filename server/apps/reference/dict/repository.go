// Package dict 数据字典仓储
package dict

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

type PostgresDictRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDictRepository(pool *pgxpool.Pool) *PostgresDictRepository {
	return &PostgresDictRepository{pool: pool}
}

// ========== 字典主表（租户级） ==========

func (r *PostgresDictRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Dict, int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
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

	args := []any{tenantID}
	where := "is_deleted = FALSE AND tenant_id = $1 AND scope = 'tenant'"
	if keyword != "" {
		where += " AND (code ILIKE $2 OR name ILIKE $2)"
		args = append(args, "%"+keyword+"%")
	}

	var total int64
	if err := q.QueryRow(ctx, "SELECT COUNT(*) FROM dicts WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count dicts: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, size, offset)
	query := `SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts
		WHERE ` + where + `
		ORDER BY sort ASC, id ASC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	rows, err := q.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list dicts: %w", err)
	}
	defer rows.Close()

	list := make([]Dict, 0)
	for rows.Next() {
		d, err := scanDict(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan dict: %w", err)
		}
		list = append(list, d)
	}
	return list, total, nil
}

func (r *PostgresDictRepository) GetByID(ctx context.Context, id uint) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts WHERE is_deleted = FALSE AND id = $1`, id)
	d, err := scanDict(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDictRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2 AND scope = 'tenant' LIMIT 1`,
		tenantID, code)
	d, err := scanDict(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDictRepository) Create(ctx context.Context, tenantID uint, req CreateDictRepoReq) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		INSERT INTO dicts (tenant_id, code, name, sort, status, scope, extend)
		VALUES ($1, $2, $3, $4, $5, 'tenant', $6::jsonb)
		RETURNING id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		tenantID, req.Code, req.Name, req.Sort, req.Status, extendJSON)
	d, err := scanDict(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDictCodeExists
		}
		return nil, fmt.Errorf("create dict: %w", err)
	}
	return &d, nil
}

func (r *PostgresDictRepository) Update(ctx context.Context, id uint, req UpdateDictRepoReq) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		UPDATE dicts SET name = $2, sort = $3, status = $4, extend = $5::jsonb, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		id, req.Name, req.Sort, req.Status, extendJSON)
	d, err := scanDict(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, fmt.Errorf("update dict: %w", err)
	}
	return &d, nil
}

func (r *PostgresDictRepository) Delete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE dicts SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete dict: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictNotFound
	}
	return nil
}

func (r *PostgresDictRepository) CountItems(ctx context.Context, dictID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return 0, err
	}
	var n int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM dict_items WHERE is_deleted = FALSE AND dict_id = $1`, dictID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return n, nil
}

// ========== 字典项（租户级） ==========

func (r *PostgresDictRepository) ListItems(ctx context.Context, dictID uint) ([]DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dict_items
		WHERE is_deleted = FALSE AND dict_id = $1
		ORDER BY sort ASC, id ASC`, dictID)
	if err != nil {
		return nil, fmt.Errorf("list dict items: %w", err)
	}
	defer rows.Close()

	list := make([]DictItem, 0)
	for rows.Next() {
		it, err := scanDictItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dict item: %w", err)
		}
		list = append(list, it)
	}
	return list, nil
}

func (r *PostgresDictRepository) GetItemByID(ctx context.Context, id uint) (*DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dict_items WHERE is_deleted = FALSE AND id = $1`, id)
	it, err := scanDictItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictItemNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *PostgresDictRepository) CreateItem(ctx context.Context, tenantID, dictID uint, req CreateDictItemRepoReq) (*DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend, is_override)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, FALSE)
		RETURNING id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		tenantID, dictID, req.Code, req.Name, req.Sort, req.Status, extendJSON)
	it, err := scanDictItem(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDictItemCodeExists
		}
		return nil, fmt.Errorf("create dict item: %w", err)
	}
	return &it, nil
}

func (r *PostgresDictRepository) UpdateItem(ctx context.Context, id uint, req UpdateDictItemRepoReq) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	tag, err := q.Exec(ctx, `
		UPDATE dict_items SET name = $2, sort = $3, status = $4, extend = $5::jsonb, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id, req.Name, req.Sort, req.Status, extendJSON)
	if err != nil {
		return fmt.Errorf("update dict item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictItemNotFound
	}
	return nil
}

func (r *PostgresDictRepository) DeleteItem(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE dict_items SET is_deleted = TRUE, updated_at = NOW() WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete dict item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictItemNotFound
	}
	return nil
}

// ============ Phase 0022: 平台字典 CRUD ============

func (r *PostgresDictRepository) ListPlatformDicts(ctx context.Context, keyword string, page, size int) ([]Dict, int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
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

	args := []any{}
	where := "is_deleted = FALSE AND scope = 'platform'"
	if keyword != "" {
		where += " AND (code ILIKE $1 OR name ILIKE $1)"
		args = append(args, "%"+keyword+"%")
	}

	var total int64
	if err := q.QueryRow(ctx, "SELECT COUNT(*) FROM dicts WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count platform dicts: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, size, offset)
	query := `SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts
		WHERE ` + where + `
		ORDER BY sort ASC, id ASC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	rows, err := q.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list platform dicts: %w", err)
	}
	defer rows.Close()

	list := make([]Dict, 0)
	for rows.Next() {
		d, err := scanDict(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan dict: %w", err)
		}
		list = append(list, d)
	}
	return list, total, nil
}

func (r *PostgresDictRepository) GetPlatformDictByID(ctx context.Context, id uint) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts WHERE is_deleted = FALSE AND id = $1 AND scope = 'platform'`, id)
	d, err := scanDict(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDictRepository) GetPlatformDictByCode(ctx context.Context, code string) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts WHERE is_deleted = FALSE AND code = $1 AND scope = 'platform' LIMIT 1`, code)
	d, err := scanDict(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *PostgresDictRepository) CreatePlatformDict(ctx context.Context, req CreateDictRepoReq) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	visibility := req.Code
	if req.Extend != nil {
		if v, ok := req.Extend["visibility"].(string); ok && v != "" {
			visibility = v
		}
	}
	if visibility == "" {
		visibility = VisibilityAll
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		INSERT INTO dicts (tenant_id, code, name, sort, status, scope, visibility, extend)
		VALUES (0, $1, $2, $3, $4, 'platform', $5, $6::jsonb)
		RETURNING id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		req.Code, req.Name, req.Sort, req.Status, visibility, extendJSON)
	d, err := scanDict(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDictCodeExists
		}
		return nil, fmt.Errorf("create platform dict: %w", err)
	}
	return &d, nil
}

func (r *PostgresDictRepository) UpdatePlatformDict(ctx context.Context, id uint, req UpdateDictRepoReq) (*Dict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		UPDATE dicts SET name = $2, sort = $3, status = $4, extend = $5::jsonb, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1 AND scope = 'platform'
		RETURNING id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		id, req.Name, req.Sort, req.Status, extendJSON)
	d, err := scanDict(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, fmt.Errorf("update platform dict: %w", err)
	}
	return &d, nil
}

func (r *PostgresDictRepository) DeletePlatformDict(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	// 删除前检查是否有租户覆盖项
	var overrideCount int64
	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM dict_items WHERE is_deleted = FALSE
		AND dict_id = $1 AND (is_override = TRUE OR platform_item_id IS NOT NULL)
	`, id).Scan(&overrideCount); err != nil {
		return fmt.Errorf("count overrides: %w", err)
	}
	if overrideCount > 0 {
		return fmt.Errorf("%w (覆盖项数=%d)", ErrPlatformItemHasOverrides, overrideCount)
	}

	tag, err := q.Exec(ctx, `UPDATE dicts SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1 AND scope = 'platform'`, id)
	if err != nil {
		return fmt.Errorf("delete platform dict: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictNotFound
	}
	return nil
}

// ============ 平台字典项 CRUD ============

func (r *PostgresDictRepository) ListPlatformItems(ctx context.Context, dictID uint) ([]DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dict_items
		WHERE is_deleted = FALSE AND dict_id = $1 AND tenant_id = 0
		ORDER BY sort ASC, id ASC`, dictID)
	if err != nil {
		return nil, fmt.Errorf("list platform items: %w", err)
	}
	defer rows.Close()

	list := make([]DictItem, 0)
	for rows.Next() {
		it, err := scanDictItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dict item: %w", err)
		}
		list = append(list, it)
	}
	return list, nil
}

func (r *PostgresDictRepository) CreatePlatformItem(ctx context.Context, dictID uint, req CreateDictItemRepoReq) (*DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend, is_override)
		VALUES (0, $1, $2, $3, $4, $5, $6::jsonb, FALSE)
		RETURNING id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		dictID, req.Code, req.Name, req.Sort, req.Status, extendJSON)
	it, err := scanDictItem(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDictItemCodeExists
		}
		return nil, fmt.Errorf("create platform item: %w", err)
	}
	return &it, nil
}

func (r *PostgresDictRepository) UpdatePlatformItem(ctx context.Context, id uint, req UpdateDictItemRepoReq) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	tag, err := q.Exec(ctx, `
		UPDATE dict_items SET name = $2, sort = $3, status = $4, extend = $5::jsonb, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1 AND tenant_id = 0`, id, req.Name, req.Sort, req.Status, extendJSON)
	if err != nil {
		return fmt.Errorf("update platform item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictItemNotFound
	}
	return nil
}

func (r *PostgresDictRepository) DeletePlatformItem(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	// 检查是否有租户覆盖
	n, err := r.CountTenantOverridesForPlatformItem(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w (覆盖租户数=%d)", ErrPlatformItemHasOverrides, n)
	}
	tag, err := q.Exec(ctx, `UPDATE dict_items SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1 AND tenant_id = 0`, id)
	if err != nil {
		return fmt.Errorf("delete platform item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictItemNotFound
	}
	return nil
}

func (r *PostgresDictRepository) CountTenantOverridesForPlatformItem(ctx context.Context, platformItemID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return 0, err
	}
	var n int64
	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM dict_items
		WHERE is_deleted = FALSE AND platform_item_id = $1 AND is_override = TRUE
	`, platformItemID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count overrides for platform item: %w", err)
	}
	return n, nil
}

// ============ dict_visibility 维护 ============

func (r *PostgresDictRepository) ListVisibilityByDict(ctx context.Context, dictID uint) ([]DictVisibility, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, dict_id, tenant_id, access, created_at, updated_at
		FROM dict_visibility WHERE dict_id = $1
		ORDER BY tenant_id ASC`, dictID)
	if err != nil {
		return nil, fmt.Errorf("list visibility: %w", err)
	}
	defer rows.Close()

	list := make([]DictVisibility, 0)
	for rows.Next() {
		var v DictVisibility
		if err := rows.Scan(&v.ID, &v.DictID, &v.TenantID, &v.Access, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *PostgresDictRepository) UpsertVisibility(ctx context.Context, dictID, tenantID uint, access string) (*DictVisibility, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		INSERT INTO dict_visibility (dict_id, tenant_id, access, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (dict_id, tenant_id) DO UPDATE
		SET access = EXCLUDED.access, updated_at = NOW()
		RETURNING id, dict_id, tenant_id, access, created_at, updated_at`,
		dictID, tenantID, access)
	var v DictVisibility
	if err := row.Scan(&v.ID, &v.DictID, &v.TenantID, &v.Access, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return nil, fmt.Errorf("upsert visibility: %w", err)
	}
	return &v, nil
}

func (r *PostgresDictRepository) DeleteVisibility(ctx context.Context, dictID, tenantID uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `DELETE FROM dict_visibility WHERE dict_id = $1 AND tenant_id = $2`, dictID, tenantID)
	if err != nil {
		return fmt.Errorf("delete visibility: %w", err)
	}
	return nil
}

func (r *PostgresDictRepository) GetAccessForTenant(ctx context.Context, dictID, tenantID uint) (string, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return "", err
	}
	var access string
	err = q.QueryRow(ctx, `SELECT access FROM dict_visibility WHERE dict_id = $1 AND tenant_id = $2`,
		dictID, tenantID).Scan(&access)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // 无显式配置
		}
		return "", fmt.Errorf("get access for tenant: %w", err)
	}
	return access, nil
}

// ============ 租户覆盖 override 维护 ============

func (r *PostgresDictRepository) GetOverrideByPlatformItem(ctx context.Context, platformItemID, tenantID uint) (*DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dict_items
		WHERE is_deleted = FALSE AND platform_item_id = $1 AND tenant_id = $2 AND is_override = TRUE
		LIMIT 1`, platformItemID, tenantID)
	it, err := scanDictItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictItemNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *PostgresDictRepository) UpsertOverride(ctx context.Context, tenantID, dictID uint, platformItemID uint, req UpdateDictItemRepoReq) (*DictItem, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)

	// 从 platform item 取 code（覆盖行必须保持 code 一致以确保唯一索引通过）
	var pCode string
	if err := q.QueryRow(ctx, `SELECT code FROM dict_items WHERE id = $1 AND tenant_id = 0`,
		platformItemID).Scan(&pCode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictItemNotFound
		}
		return nil, fmt.Errorf("get platform item code: %w", err)
	}

	row := q.QueryRow(ctx, `
		INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend, platform_item_id, is_override)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, TRUE)
		ON CONFLICT (tenant_id, platform_item_id) WHERE is_override = TRUE AND is_deleted = FALSE
		DO UPDATE SET name = EXCLUDED.name, sort = EXCLUDED.sort, status = EXCLUDED.status,
			extend = EXCLUDED.extend, updated_at = NOW()
		RETURNING id, tenant_id, dict_id, code, name, sort, status, platform_item_id, is_override,
		COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		tenantID, dictID, pCode, req.Name, req.Sort, req.Status, extendJSON, platformItemID)
	it, err := scanDictItem(row)
	if err != nil {
		return nil, fmt.Errorf("upsert override: %w", err)
	}
	return &it, nil
}

func (r *PostgresDictRepository) DeleteOverride(ctx context.Context, tenantID, platformItemID uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `UPDATE dict_items SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND tenant_id = $1 AND platform_item_id = $2 AND is_override = TRUE`,
		tenantID, platformItemID)
	if err != nil {
		return fmt.Errorf("delete override: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDictItemNotFound
	}
	return nil
}

// ============ Resolve 合并查询（业务最终消费） ============

// ResolveDictForTenant 按 code 合并字典（租户级优先 / 平台级 + 覆盖）
func (r *PostgresDictRepository) ResolveDictForTenant(ctx context.Context, tenantID uint, dictCode string) (*ResolvedDict, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}

	// 1) 优先租户自建字典
	var td Dict
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts
		WHERE is_deleted = FALSE AND scope = 'tenant' AND tenant_id = $1 AND code = $2
		LIMIT 1`, tenantID, dictCode).Scan(
		&td.ID, &td.TenantID, &td.Code, &td.Name, &td.Sort, &td.Status, &td.Scope, &td.Visibility,
		&td.Extend, &td.CreatedAt, &td.UpdatedAt)
	hasTenantDict := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("lookup tenant dict: %w", err)
	}

	if hasTenantDict {
		// 租户自建字典：纯租户项
		items, err := r.loadTenantOwnedItems(ctx, td.ID, tenantID)
		if err != nil {
			return nil, err
		}
		return &ResolvedDict{
			DictID:      td.ID,
			Code:        td.Code,
			Name:        td.Name,
			Scope:       ScopeTenant,
			Access:      AccessOwned,
			HasOverride: false,
			Items:       items,
		}, nil
	}

	// 2) 走平台字典路径
	var pd Dict
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, scope, visibility,
		COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts
		WHERE is_deleted = FALSE AND scope = 'platform' AND code = $1
		LIMIT 1`, dictCode).Scan(
		&pd.ID, &pd.TenantID, &pd.Code, &pd.Name, &pd.Sort, &pd.Status, &pd.Scope, &pd.Visibility,
		&pd.Extend, &pd.CreatedAt, &pd.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDictNotFound
		}
		return nil, fmt.Errorf("lookup platform dict: %w", err)
	}

	// 3) 计算 access
	access, err := r.resolveAccess(ctx, pd, tenantID)
	if err != nil {
		return nil, err
	}
	if access == AccessInvisible {
		return nil, ErrDictInvisible
	}

	// 4) 合并字典项
	items, hasOverride, err := r.mergeItemsForTenant(ctx, pd.ID, tenantID, access)
	if err != nil {
		return nil, err
	}

	return &ResolvedDict{
		DictID:      pd.ID,
		Code:        pd.Code,
		Name:        pd.Name,
		Scope:       ScopePlatform,
		Access:      access,
		HasOverride: hasOverride,
		Items:       items,
	}, nil
}

func (r *PostgresDictRepository) ResolveDictByIDForTenant(ctx context.Context, tenantID, dictID uint) (*ResolvedDict, error) {
	d, err := r.GetByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	if d.Scope == ScopeTenant && d.TenantID != tenantID {
		return nil, ErrDictNotFound
	}
	return r.ResolveDictForTenant(ctx, tenantID, d.Code)
}

// resolveAccess 解析平台字典对租户的 access
func (r *PostgresDictRepository) resolveAccess(ctx context.Context, d Dict, tenantID uint) (string, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return "", err
	}
	// 先查显式配置
	explicit, err := r.GetAccessForTenant(ctx, d.ID, tenantID)
	if err != nil {
		return "", err
	}
	if explicit != "" {
		return explicit, nil
	}

	// 没显式配置按 visibility 默认策略
	switch d.Visibility {
	case VisibilityWhitelist:
		// 白名单模式：未命中即不可见
		return AccessInvisible, nil
	case VisibilityBlacklist:
		// 黑名单模式：未命中即可编辑
		return AccessEditable, nil
	default: // VisibilityAll
		// 默认所有租户可编辑
		_ = q
		return AccessEditable, nil
	}
}

// mergeItemsForTenant 合并 platform items + 租户覆盖
func (r *PostgresDictRepository) mergeItemsForTenant(ctx context.Context, dictID, tenantID uint, access string) ([]ResolvedItem, bool, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, false, err
	}

	// editable 时合并 overrides；readonly 时只看平台项
	if access != AccessEditable {
		items, err := r.ListPlatformItems(ctx, dictID)
		if err != nil {
			return nil, false, err
		}
		out := make([]ResolvedItem, 0, len(items))
		for _, it := range items {
			out = append(out, ResolvedItem{
				ItemID:         it.ID,
				PlatformItemID: it.ID,
				Code:           it.Code,
				Name:           it.Name,
				Sort:           it.Sort,
				IsOverride:     false,
			})
		}
		return out, false, nil
	}

	// editable: COALESCE 合并
	rows, err := q.Query(ctx, `
		SELECT
			COALESCE(o.id, p.id)              AS item_id,
			p.id                              AS platform_item_id,
			COALESCE(o.code, p.code)          AS code,
			COALESCE(o.name, p.name)          AS name,
			COALESCE(o.sort, p.sort)          AS sort,
			(o.id IS NOT NULL)                AS is_override
		FROM dict_items p
		LEFT JOIN dict_items o
			ON o.platform_item_id = p.id
		   AND o.tenant_id = $2
		   AND o.is_override = TRUE
		   AND o.is_deleted = FALSE
		WHERE p.dict_id = $1
		  AND p.tenant_id = 0
		  AND p.is_deleted = FALSE
		ORDER BY COALESCE(o.sort, p.sort) ASC, p.id ASC
	`, dictID, tenantID)
	if err != nil {
		return nil, false, fmt.Errorf("merge items: %w", err)
	}
	defer rows.Close()

	out := make([]ResolvedItem, 0)
	hasOverride := false
	for rows.Next() {
		var ri ResolvedItem
		var isOverride bool
		if err := rows.Scan(&ri.ItemID, &ri.PlatformItemID, &ri.Code, &ri.Name, &ri.Sort, &isOverride); err != nil {
			return nil, false, fmt.Errorf("scan resolved item: %w", err)
		}
		ri.IsOverride = isOverride
		if isOverride {
			hasOverride = true
		}
		out = append(out, ri)
	}
	return out, hasOverride, nil
}

// loadTenantOwnedItems 加载租户自建字典的项
func (r *PostgresDictRepository) loadTenantOwnedItems(ctx context.Context, dictID, tenantID uint) ([]ResolvedItem, error) {
	items, err := r.ListItems(ctx, dictID)
	if err != nil {
		return nil, err
	}
	_ = tenantID
	out := make([]ResolvedItem, 0, len(items))
	for _, it := range items {
		out = append(out, ResolvedItem{
			ItemID:         it.ID,
			PlatformItemID: 0,
			Code:           it.Code,
			Name:           it.Name,
			Sort:           it.Sort,
			IsOverride:     false,
		})
	}
	return out, nil
}

// ========== helpers ==========

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDict(row rowScanner) (Dict, error) {
	var d Dict
	var extendJSON []byte
	if err := row.Scan(
		&d.ID, &d.TenantID, &d.Code, &d.Name, &d.Sort, &d.Status, &d.Scope, &d.Visibility,
		&extendJSON, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		return Dict{}, err
	}
	if len(extendJSON) > 0 {
		_ = json.Unmarshal(extendJSON, &d.Extend)
	}
	return d, nil
}

func scanDictItem(row rowScanner) (DictItem, error) {
	var it DictItem
	var extendJSON []byte
	if err := row.Scan(
		&it.ID, &it.TenantID, &it.DictID, &it.Code, &it.Name, &it.Sort, &it.Status,
		&it.PlatformItemID, &it.IsOverride,
		&extendJSON, &it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		return DictItem{}, err
	}
	if len(extendJSON) > 0 {
		_ = json.Unmarshal(extendJSON, &it.Extend)
	}
	return it, nil
}

// isUniqueViolation 简化判断 pgx 唯一键冲突（SQLSTATE 23505）
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

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

func (r *PostgresDictRepository) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Dict, int64, error) {
	q, err := db.GetQuerier(ctx)
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
	where := "is_deleted = FALSE AND tenant_id = $1"
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
	query := `SELECT id, tenant_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at
		FROM dicts WHERE is_deleted = FALSE AND tenant_id = $1 AND code = $2`, tenantID, code)
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		INSERT INTO dicts (tenant_id, code, name, sort, status, extend)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at`,
		tenantID, req.Code, req.Name, req.Sort, req.Status, extendJSON)
	d, err := scanDict(row)
	if err != nil {
		// 唯一键冲突
		if isUniqueViolation(err) {
			return nil, ErrDictCodeExists
		}
		return nil, fmt.Errorf("create dict: %w", err)
	}
	return &d, nil
}

func (r *PostgresDictRepository) Update(ctx context.Context, id uint, req UpdateDictRepoReq) (*Dict, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		UPDATE dicts SET name = $2, sort = $3, status = $4, extend = $5, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1
		RETURNING id, tenant_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at`,
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
	q, err := db.GetQuerier(ctx)
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return 0, err
	}
	var n int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM dict_items WHERE is_deleted = FALSE AND dict_id = $1`, dictID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count dict items: %w", err)
	}
	return n, nil
}

// ========== 字典项 ==========

func (r *PostgresDictRepository) ListItems(ctx context.Context, dictID uint) ([]DictItem, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, dict_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	row := q.QueryRow(ctx, `
		SELECT id, tenant_id, dict_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	row := q.QueryRow(ctx, `
		INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status, extend)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, dict_id, code, name, sort, status, COALESCE(extend, '{}') AS extend, created_at, updated_at`,
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
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	extendJSON, _ := json.Marshal(req.Extend)
	tag, err := q.Exec(ctx, `
		UPDATE dict_items SET name = $2, sort = $3, status = $4, extend = $5, updated_at = NOW()
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
	q, err := db.GetQuerier(ctx)
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

// ========== helpers ==========

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDict(row rowScanner) (Dict, error) {
	var d Dict
	var extendJSON []byte
	if err := row.Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &d.Sort, &d.Status, &extendJSON, &d.CreatedAt, &d.UpdatedAt); err != nil {
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
	if err := row.Scan(&it.ID, &it.TenantID, &it.DictID, &it.Code, &it.Name, &it.Sort, &it.Status, &extendJSON, &it.CreatedAt, &it.UpdatedAt); err != nil {
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

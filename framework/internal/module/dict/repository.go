package dict

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	dictpkg "gx1727.com/xin/framework/pkg/dict"
)

type DictRepository struct {
	db *pgxpool.Pool
}

func NewDictRepository(db *pgxpool.Pool) *DictRepository {
	return &DictRepository{db: db}
}

type DictCreate struct {
	Code   string
	Name   string
	Extend map[string]interface{}
}

func (r *DictRepository) Create(ctx context.Context, tenantID uint, req DictCreate) (*dictpkg.Dict, error) {
	extendJSON, _ := json.Marshal(req.Extend)
	var d dictpkg.Dict
	err := r.db.QueryRow(ctx, `
		INSERT INTO dicts (tenant_id, code, name, extend)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, code, name, extend
	`, tenantID, req.Code, req.Name, extendJSON).Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &extendJSON)
	if err != nil {
		return nil, fmt.Errorf("create dict: %w", err)
	}
	if extendJSON != nil {
		json.Unmarshal(extendJSON, &d.Extend)
	}
	dictpkg.Invalidate(tenantID, "")
	return &d, nil
}

func (r *DictRepository) Update(ctx context.Context, tenantID uint, id uint, name string, extend map[string]interface{}) error {
	extendJSON, _ := json.Marshal(extend)
	_, err := r.db.Exec(ctx, `
		UPDATE dicts SET name = $1, extend = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4 AND is_deleted = FALSE
	`, name, extendJSON, id, tenantID)
	if err != nil {
		return fmt.Errorf("update dict: %w", err)
	}
	dictpkg.Invalidate(tenantID, "")
	return nil
}

func (r *DictRepository) Delete(ctx context.Context, tenantID uint, id uint) error {
	_, err := r.db.Exec(ctx, `
		UPDATE dicts SET is_deleted = TRUE, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete dict: %w", err)
	}
	dictpkg.Invalidate(tenantID, "")
	return nil
}

type DictItemCreate struct {
	Code   string
	Name   string
	Sort   int
	Extend map[string]interface{}
}

func (r *DictRepository) CreateItem(ctx context.Context, tenantID uint, dictID uint, req DictItemCreate) (*dictpkg.DictItem, error) {
	extendJSON, _ := json.Marshal(req.Extend)
	var item dictpkg.DictItem
	err := r.db.QueryRow(ctx, `
		INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, extend)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, code, name, sort, extend
	`, tenantID, dictID, req.Code, req.Name, req.Sort, extendJSON).Scan(&item.ID, &item.Code, &item.Name, &item.Sort, &extendJSON)
	if err != nil {
		return nil, fmt.Errorf("create dict item: %w", err)
	}
	if extendJSON != nil {
		json.Unmarshal(extendJSON, &item.Extend)
	}
	dictpkg.Invalidate(tenantID, "")
	return &item, nil
}

func (r *DictRepository) UpdateItem(ctx context.Context, tenantID uint, id uint, name string, sort int, extend map[string]interface{}) error {
	extendJSON, _ := json.Marshal(extend)
	_, err := r.db.Exec(ctx, `
		UPDATE dict_items SET name = $1, sort = $2, extend = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5 AND is_deleted = FALSE
	`, name, sort, extendJSON, id, tenantID)
	if err != nil {
		return fmt.Errorf("update dict item: %w", err)
	}
	dictpkg.Invalidate(tenantID, "")
	return nil
}

func (r *DictRepository) DeleteItem(ctx context.Context, tenantID uint, id uint) error {
	_, err := r.db.Exec(ctx, `
		UPDATE dict_items SET is_deleted = TRUE, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete dict item: %w", err)
	}
	dictpkg.Invalidate(tenantID, "")
	return nil
}

func (r *DictRepository) List(ctx context.Context, tenantID uint) ([]dictpkg.Dict, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, code, name, extend
		FROM dicts
		WHERE tenant_id = $1 AND is_deleted = FALSE
		ORDER BY code
	`, tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("list dicts: %w", err)
	}
	defer rows.Close()

	var list []dictpkg.Dict
	for rows.Next() {
		var d dictpkg.Dict
		var extendJSON []byte
		err := rows.Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &extendJSON)
		if err != nil {
			return nil, 0, fmt.Errorf("scan dict: %w", err)
		}
		if extendJSON != nil {
			json.Unmarshal(extendJSON, &d.Extend)
		}
		list = append(list, d)
	}

	var total int64
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM dicts WHERE tenant_id = $1 AND is_deleted = FALSE", tenantID).Scan(&total)

	return list, total, nil
}

func (r *DictRepository) GetByCode(ctx context.Context, tenantID uint, code string) (*dictpkg.Dict, error) {
	var d dictpkg.Dict
	var extendJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, extend
		FROM dicts
		WHERE tenant_id = $1 AND code = $2 AND is_deleted = FALSE
	`, tenantID, code).Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &extendJSON)
	if err != nil {
		return nil, fmt.Errorf("get dict: %w", err)
	}
	if extendJSON != nil {
		json.Unmarshal(extendJSON, &d.Extend)
	}
	return &d, nil
}

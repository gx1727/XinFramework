package dict

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DictItem struct {
	ID     uint                   `json:"id"`
	Code   string                 `json:"code"`
	Name   string                 `json:"name"`
	Sort   int                    `json:"sort"`
	Extend map[string]interface{} `json:"extend,omitempty"`
}

type Dict struct {
	ID       uint                   `json:"id"`
	TenantID uint                   `json:"tenant_id"`
	Code     string                 `json:"code"`
	Name     string                 `json:"name"`
	Extend   map[string]interface{} `json:"extend,omitempty"`
	Items    []DictItem             `json:"items"`
}

type Cache struct {
	mu sync.RWMutex
	// tenantID -> dictCode -> Dict
	// 缓存的是"该租户视角下"看到的最终 Dict（含平台项 + 租户覆盖项）
	data map[uint]map[string]*Dict
}

var (
	globalCache = &Cache{
		data: make(map[uint]map[string]*Dict),
	}
	dbPool *pgxpool.Pool
)

func Init(pool *pgxpool.Pool) error {
	dbPool = pool
	return nil
}

func GetPool() *pgxpool.Pool {
	return dbPool
}

// LoadTenant 加载（平台 + 租户）的字典到缓存，按 code 项级合并：租户项覆盖平台项
func LoadTenant(ctx context.Context, tenantID uint) error {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	if globalCache.data[tenantID] == nil {
		globalCache.data[tenantID] = make(map[string]*Dict)
	} else {
		globalCache.data[tenantID] = make(map[string]*Dict)
	}

	rows, err := dbPool.Query(ctx, `
		SELECT d.id, d.tenant_id, d.code, d.name,
		       COALESCE(d.extend, '{}') as extend,
		       COALESCE(di.id, 0) as item_id,
		       COALESCE(di.code, '') as item_code,
		       COALESCE(di.name, '') as item_name,
		       COALESCE(di.sort, 0) as item_sort,
		       COALESCE(di.extend, '{}') as item_extend
		FROM (SELECT * FROM dicts WHERE tenant_id IN (0, $1) AND is_deleted = FALSE) d
		LEFT JOIN dict_items di ON di.dict_id = d.id AND di.is_deleted = FALSE
		ORDER BY d.code, d.tenant_id ASC, item_sort, item_id
	`, tenantID)
	if err != nil {
		return fmt.Errorf("load dicts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d struct {
			ID, TenantID uint
			Code, Name   string
			Extend       []byte
		}
		var item struct {
			ID, Sort   uint
			Code, Name string
			Extend     []byte
		}

		if err := rows.Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &d.Extend, &item.ID, &item.Code, &item.Name, &item.Sort, &item.Extend); err != nil {
			return fmt.Errorf("scan dict row: %w", err)
		}

		merged := globalCache.data[tenantID][d.Code]
		if merged == nil {
			merged = &Dict{
				ID:       d.ID,
				TenantID: d.TenantID,
				Code:     d.Code,
				Name:     d.Name,
				Items:    []DictItem{},
			}
			if len(d.Extend) > 0 {
				_ = json.Unmarshal(d.Extend, &merged.Extend)
			}
			globalCache.data[tenantID][d.Code] = merged
		} else {
			// 后续出现的 dict 行：租户的覆盖平台的元数据
			if d.TenantID != 0 {
				merged.ID = d.ID
				merged.TenantID = d.TenantID
				merged.Name = d.Name
				if len(d.Extend) > 0 {
					_ = json.Unmarshal(d.Extend, &merged.Extend)
				}
			}
		}

		if item.ID > 0 {
			merged.Items = append(merged.Items, DictItem{
				ID:     item.ID,
				Code:   item.Code,
				Name:   item.Name,
				Sort:   int(item.Sort),
				Extend: nil,
			})
			if len(item.Extend) > 0 {
				last := &merged.Items[len(merged.Items)-1]
				_ = json.Unmarshal(item.Extend, &last.Extend)
			}
		}
	}

	// 同一 code 下的 items 按 code 去重：后者（租户）覆盖前者（平台）
	for _, merged := range globalCache.data[tenantID] {
		merged.Items = dedupItemsByCode(merged.Items)
	}

	return nil
}

// dedupItemsByCode 按 code 去重，保留最后一次出现（租户覆盖平台）
func dedupItemsByCode(items []DictItem) []DictItem {
	if len(items) <= 1 {
		return items
	}
	idxByCode := make(map[string]int, len(items))
	out := make([]DictItem, 0, len(items))
	for _, it := range items {
		if existing, ok := idxByCode[it.Code]; ok {
			out[existing] = it
		} else {
			idxByCode[it.Code] = len(out)
			out = append(out, it)
		}
	}
	return out
}

func Get(tenantID uint, dictCode string) (*Dict, bool) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()

	if globalCache.data[tenantID] == nil {
		return nil, false
	}
	d, ok := globalCache.data[tenantID][dictCode]
	return d, ok
}

func GetItem(tenantID uint, dictCode string, itemCode string) (*DictItem, bool) {
	d, ok := Get(tenantID, dictCode)
	if !ok {
		return nil, false
	}
	for _, item := range d.Items {
		if item.Code == itemCode {
			return &item, true
		}
	}
	return nil, false
}

func GetItems(tenantID uint, dictCode string) []DictItem {
	d, ok := Get(tenantID, dictCode)
	if !ok {
		return nil
	}
	return d.Items
}

func RefreshTenant(ctx context.Context, tenantID uint) error {
	globalCache.mu.Lock()
	delete(globalCache.data, tenantID)
	globalCache.mu.Unlock()

	return LoadTenant(ctx, tenantID)
}

// RefreshDict 重新加载（平台 + 租户）该 code 字典到该租户的缓存槽
// 当前所有调用都是租户改自己字典的场景；平台字典修改尚未开放
func RefreshDict(ctx context.Context, tenantID uint, dictCode string) error {
	globalCache.mu.Lock()
	if globalCache.data[tenantID] != nil {
		delete(globalCache.data[tenantID], dictCode)
	}
	globalCache.mu.Unlock()

	rows, err := dbPool.Query(ctx, `
		SELECT d.id, d.tenant_id, d.code, d.name,
		       COALESCE(d.extend, '{}') as extend,
		       COALESCE(di.id, 0) as item_id,
		       COALESCE(di.code, '') as item_code,
		       COALESCE(di.name, '') as item_name,
		       COALESCE(di.sort, 0) as item_sort,
		       COALESCE(di.extend, '{}') as item_extend
		FROM (SELECT * FROM dicts WHERE tenant_id IN (0, $1) AND code = $2 AND is_deleted = FALSE) d
		LEFT JOIN dict_items di ON di.dict_id = d.id AND di.is_deleted = FALSE
		ORDER BY d.tenant_id ASC, item_sort, item_id
	`, tenantID, dictCode)
	if err != nil {
		return fmt.Errorf("refresh dict query: %w", err)
	}
	defer rows.Close()

	var merged *Dict
	for rows.Next() {
		var d struct {
			ID, TenantID uint
			Code, Name   string
			Extend       []byte
		}
		var item struct {
			ID, Sort   uint
			Code, Name string
			Extend     []byte
		}

		if err := rows.Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &d.Extend, &item.ID, &item.Code, &item.Name, &item.Sort, &item.Extend); err != nil {
			return fmt.Errorf("refresh scan: %w", err)
		}

		if merged == nil {
			merged = &Dict{
				ID:       d.ID,
				TenantID: d.TenantID,
				Code:     d.Code,
				Name:     d.Name,
				Items:    []DictItem{},
			}
			if len(d.Extend) > 0 {
				_ = json.Unmarshal(d.Extend, &merged.Extend)
			}
		} else if d.TenantID != 0 {
			// 租户的 dict 元数据覆盖平台
			merged.ID = d.ID
			merged.TenantID = d.TenantID
			merged.Name = d.Name
			if len(d.Extend) > 0 {
				_ = json.Unmarshal(d.Extend, &merged.Extend)
			}
		}

		if item.ID > 0 {
			merged.Items = append(merged.Items, DictItem{
				ID:     item.ID,
				Code:   item.Code,
				Name:   item.Name,
				Sort:   int(item.Sort),
				Extend: nil,
			})
			if len(item.Extend) > 0 {
				last := &merged.Items[len(merged.Items)-1]
				_ = json.Unmarshal(item.Extend, &last.Extend)
			}
		}
	}

	if merged != nil {
		merged.Items = dedupItemsByCode(merged.Items)
		globalCache.mu.Lock()
		if globalCache.data[tenantID] == nil {
			globalCache.data[tenantID] = make(map[string]*Dict)
		}
		globalCache.data[tenantID][dictCode] = merged
		globalCache.mu.Unlock()
	}

	return nil
}

func Invalidate(tenantID uint, dictCode string) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	if globalCache.data[tenantID] != nil {
		delete(globalCache.data[tenantID], dictCode)
	}
}

func InvalidateTenant(tenantID uint) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	delete(globalCache.data, tenantID)
}

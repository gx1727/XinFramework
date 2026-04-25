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
		FROM (SELECT * FROM dicts WHERE tenant_id = $1 AND is_deleted = FALSE) d
		LEFT JOIN dict_items di ON di.dict_id = d.id AND di.is_deleted = FALSE
		ORDER BY d.code, item_sort, item_id
	`, tenantID)
	if err != nil {
		return fmt.Errorf("load dicts: %w", err)
	}
	defer rows.Close()

	currentCode := ""
	currentDict := (*Dict)(nil)

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

		err := rows.Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &d.Extend, &item.ID, &item.Code, &item.Name, &item.Sort, &item.Extend)
		if err != nil {
			return fmt.Errorf("scan dict row: %w", err)
		}

		if d.Code != currentCode {
			if currentDict != nil {
				globalCache.data[tenantID][currentCode] = currentDict
			}
			currentDict = &Dict{
				ID:       d.ID,
				TenantID: d.TenantID,
				Code:     d.Code,
				Name:     d.Name,
				Items:    []DictItem{},
			}
			if len(d.Extend) > 0 {
				json.Unmarshal(d.Extend, &currentDict.Extend)
			}
			currentCode = d.Code
		}

		if item.ID > 0 {
			di := DictItem{ID: item.ID, Code: item.Code, Name: item.Name, Sort: int(item.Sort)}
			if len(item.Extend) > 0 {
				json.Unmarshal(item.Extend, &di.Extend)
			}
			currentDict.Items = append(currentDict.Items, di)
		}
	}

	if currentDict != nil {
		globalCache.data[tenantID][currentCode] = currentDict
	}

	return nil
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

func RefreshDict(ctx context.Context, tenantID uint, dictCode string) error {
	globalCache.mu.Lock()
	delete(globalCache.data[tenantID], dictCode)
	globalCache.mu.Unlock()

	rows, err := dbPool.Query(ctx, `
		SELECT d.id, d.tenant_id, d.code, d.name,
		       COALESCE(d.extend, '{}') as extend,
		       COALESCE(di.id, 0) as item_id,
		       COALESCE(di.code, '') as item_code,
		       COALESCE(di.name, '') as item_name,
		       COALESCE(di.sort, 0) as item_sort,
		       COALESCE(di.extend, '{}') as item_extend
		FROM (SELECT * FROM dicts WHERE tenant_id = $1 AND code = $2 AND is_deleted = FALSE) d
		LEFT JOIN dict_items di ON di.dict_id = d.id AND di.is_deleted = FALSE
		ORDER BY item_sort, item_id
	`, tenantID, dictCode)
	if err != nil {
		return fmt.Errorf("refresh dict query: %w", err)
	}
	defer rows.Close()

	var d *Dict
	for rows.Next() {
		if d == nil {
			d = &Dict{Items: []DictItem{}}
		}
		var dd struct {
			ID, TenantID uint
			Code, Name   string
			Extend       []byte
		}
		var item struct {
			ID, Sort   uint
			Code, Name string
			Extend     []byte
		}

		err := rows.Scan(&dd.ID, &dd.TenantID, &dd.Code, &dd.Name, &dd.Extend, &item.ID, &item.Code, &item.Name, &item.Sort, &item.Extend)
		if err != nil {
			return fmt.Errorf("refresh scan: %w", err)
		}

		if d.Code == "" {
			d.ID = dd.ID
			d.TenantID = dd.TenantID
			d.Code = dd.Code
			d.Name = dd.Name
			if len(dd.Extend) > 0 {
				json.Unmarshal(dd.Extend, &d.Extend)
			}
		}

		if item.ID > 0 {
			di := DictItem{ID: item.ID, Code: item.Code, Name: item.Name, Sort: int(item.Sort)}
			if len(item.Extend) > 0 {
				json.Unmarshal(item.Extend, &di.Extend)
			}
			d.Items = append(d.Items, di)
		}
	}

	if d != nil {
		globalCache.mu.Lock()
		if globalCache.data[tenantID] == nil {
			globalCache.data[tenantID] = make(map[string]*Dict)
		}
		globalCache.data[tenantID][dictCode] = d
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

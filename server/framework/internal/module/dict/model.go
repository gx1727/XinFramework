// Package dict ????
package dict

import (
	"context"
	"time"
)

// Dict ????
type Dict struct {
	ID        uint                   `json:"id"`
	TenantID  uint                   `json:"tenant_id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	Sort      int                    `json:"sort"`
	Status    int8                   `json:"status"`
	Extend    map[string]interface{} `json:"extend,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// DictItem ???
type DictItem struct {
	ID        uint                   `json:"id"`
	TenantID  uint                   `json:"tenant_id"`
	DictID    uint                   `json:"dict_id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	Sort      int                    `json:"sort"`
	Status    int8                   `json:"status"`
	Extend    map[string]interface{} `json:"extend,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// DictRepository ??????
type DictRepository interface {
	// ????
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Dict, int64, error)
	GetByID(ctx context.Context, id uint) (*Dict, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Dict, error)
	Create(ctx context.Context, tenantID uint, req CreateDictRepoReq) (*Dict, error)
	Update(ctx context.Context, id uint, req UpdateDictRepoReq) (*Dict, error)
	Delete(ctx context.Context, id uint) error
	CountItems(ctx context.Context, dictID uint) (int64, error)

	// ???
	ListItems(ctx context.Context, dictID uint) ([]DictItem, error)
	GetItemByID(ctx context.Context, id uint) (*DictItem, error)
	CreateItem(ctx context.Context, tenantID, dictID uint, req CreateDictItemRepoReq) (*DictItem, error)
	UpdateItem(ctx context.Context, id uint, req UpdateDictItemRepoReq) error
	DeleteItem(ctx context.Context, id uint) error
}

// CreateDictRepoReq ?????????
type CreateDictRepoReq struct {
	Code   string
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

// UpdateDictRepoReq ?????????
type UpdateDictRepoReq struct {
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

// CreateDictItemRepoReq ??????????
type CreateDictItemRepoReq struct {
	Code   string
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

// UpdateDictItemRepoReq ??????????
type UpdateDictItemRepoReq struct {
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

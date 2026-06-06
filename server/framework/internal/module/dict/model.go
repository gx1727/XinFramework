// Package dict 数据字典 - 领域模型与仓储接口
package dict

import (
	"context"
	"time"
)

// Dict 字典主表
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

// DictItem 字典项
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

// DictRepository 字典仓储接口
type DictRepository interface {
	// 字典主表
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Dict, int64, error)
	GetByID(ctx context.Context, id uint) (*Dict, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Dict, error)
	Create(ctx context.Context, tenantID uint, req CreateDictRepoReq) (*Dict, error)
	Update(ctx context.Context, id uint, req UpdateDictRepoReq) (*Dict, error)
	Delete(ctx context.Context, id uint) error
	CountItems(ctx context.Context, dictID uint) (int64, error)

	// 字典项
	ListItems(ctx context.Context, dictID uint) ([]DictItem, error)
	GetItemByID(ctx context.Context, id uint) (*DictItem, error)
	CreateItem(ctx context.Context, tenantID, dictID uint, req CreateDictItemRepoReq) (*DictItem, error)
	UpdateItem(ctx context.Context, id uint, req UpdateDictItemRepoReq) error
	DeleteItem(ctx context.Context, id uint) error
}

// CreateDictRepoReq 仓储层创建字典参数
type CreateDictRepoReq struct {
	Code   string
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

// UpdateDictRepoReq 仓储层更新字典参数
type UpdateDictRepoReq struct {
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

// CreateDictItemRepoReq 仓储层创建字典项参数
type CreateDictItemRepoReq struct {
	Code   string
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

// UpdateDictItemRepoReq 仓储层更新字典项参数
type UpdateDictItemRepoReq struct {
	Name   string
	Sort   int
	Status int8
	Extend map[string]interface{}
}

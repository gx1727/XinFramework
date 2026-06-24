// Package dict 数据字典 - 领域模型与仓储接口
package dict

import (
	"context"
	"errors"
	"time"
)

// Dict 字典主表
type Dict struct {
	ID         uint                   `json:"id"`
	TenantID   uint                   `json:"tenant_id"`
	Code       string                 `json:"code"`
	Name       string                 `json:"name"`
	Sort       int                    `json:"sort"`
	Status     int8                   `json:"status"`
	Scope      string                 `json:"scope"`      // 'platform' | 'tenant'
	Visibility string                 `json:"visibility"` // 'all' | 'whitelist' | 'blacklist'
	Extend     map[string]any `json:"extend,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// DictItem 字典项
type DictItem struct {
	ID              uint                   `json:"id"`
	TenantID        uint                   `json:"tenant_id"`
	DictID          uint                   `json:"dict_id"`
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	Sort            int                    `json:"sort"`
	Status          int8                   `json:"status"`
	PlatformItemID  *uint                  `json:"platform_item_id,omitempty"` // 指向被覆盖的 platform item id
	IsOverride      bool                   `json:"is_override"`                // TRUE 表示这是租户对 platform item 的覆盖
	Extend          map[string]any `json:"extend,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// DictVisibility 平台字典对租户的访问级别
type DictVisibility struct {
	ID        uint      `json:"id"`
	DictID    uint      `json:"dict_id"`
	TenantID  uint      `json:"tenant_id"`
	Access    string    `json:"access"` // 'invisible' | 'readonly' | 'editable'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResolvedDict 业务最终消费的合并字典
// 平台字典 + 租户覆盖（COALESCE 合并）
type ResolvedDict struct {
	DictID      uint           `json:"dict_id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Scope       string         `json:"scope"`  // 'platform' | 'tenant'
	Access      string         `json:"access"` // 'readonly' | 'editable' | 'tenant_owned'
	HasOverride bool           `json:"has_override"`
	Items       []ResolvedItem `json:"items"`
}

// ResolvedItem 合并后的字典项
type ResolvedItem struct {
	ItemID         uint   `json:"item_id"`        // 当前实际返回的 item id（可能是 platform item id 或 override id）
	PlatformItemID uint   `json:"platform_item_id,omitempty"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Sort           int    `json:"sort"`
	IsOverride     bool   `json:"is_override"` // TRUE 表示来自租户覆盖
}

// 字典 scope 常量
const (
	ScopePlatform = "platform"
	ScopeTenant   = "tenant"
)

// 字典 visibility 常量
const (
	VisibilityAll       = "all"
	VisibilityWhitelist = "whitelist"
	VisibilityBlacklist = "blacklist"
)

// dict_visibility.access 常量
const (
	AccessInvisible = "invisible"
	AccessReadonly  = "readonly"
	AccessEditable  = "editable"
	AccessOwned     = "tenant_owned" // 租户自建字典的 access 标记
)

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

	// 平台字典 CRUD（跨租户，绕过 RLS）
	ListPlatformDicts(ctx context.Context, keyword string, page, size int) ([]Dict, int64, error)
	GetPlatformDictByID(ctx context.Context, id uint) (*Dict, error)
	GetPlatformDictByCode(ctx context.Context, code string) (*Dict, error)
	CreatePlatformDict(ctx context.Context, req CreateDictRepoReq) (*Dict, error)
	UpdatePlatformDict(ctx context.Context, id uint, req UpdateDictRepoReq) (*Dict, error)
	DeletePlatformDict(ctx context.Context, id uint) error

	// 平台字典项 CRUD
	ListPlatformItems(ctx context.Context, dictID uint) ([]DictItem, error)
	CreatePlatformItem(ctx context.Context, dictID uint, req CreateDictItemRepoReq) (*DictItem, error)
	UpdatePlatformItem(ctx context.Context, id uint, req UpdateDictItemRepoReq) error
	DeletePlatformItem(ctx context.Context, id uint) error
	CountTenantOverridesForPlatformItem(ctx context.Context, platformItemID uint) (int64, error)

	// dict_visibility 维护
	ListVisibilityByDict(ctx context.Context, dictID uint) ([]DictVisibility, error)
	UpsertVisibility(ctx context.Context, dictID, tenantID uint, access string) (*DictVisibility, error)
	DeleteVisibility(ctx context.Context, dictID, tenantID uint) error
	GetAccessForTenant(ctx context.Context, dictID, tenantID uint) (string, error)

	// 租户覆盖（override）维护
	GetOverrideByPlatformItem(ctx context.Context, platformItemID, tenantID uint) (*DictItem, error)
	UpsertOverride(ctx context.Context, tenantID, dictID uint, platformItemID uint, req UpdateDictItemRepoReq) (*DictItem, error)
	DeleteOverride(ctx context.Context, tenantID, platformItemID uint) error

	// Resolve 合并查询（业务最终消费）
	ResolveDictForTenant(ctx context.Context, tenantID uint, dictCode string) (*ResolvedDict, error)
	ResolveDictByIDForTenant(ctx context.Context, tenantID, dictID uint) (*ResolvedDict, error)
}

// CreateDictRepoReq 仓储层创建字典参数
type CreateDictRepoReq struct {
	Code   string
	Name   string
	Sort   int
	Status int8
	Extend map[string]any
}

// UpdateDictRepoReq 仓储层更新字典参数
type UpdateDictRepoReq struct {
	Name   string
	Sort   int
	Status int8
	Extend map[string]any
}

// CreateDictItemRepoReq 仓储层创建字典项参数
type CreateDictItemRepoReq struct {
	Code   string
	Name   string
	Sort   int
	Status int8
	Extend map[string]any
}

// UpdateDictItemRepoReq 仓储层更新字典项参数
type UpdateDictItemRepoReq struct {
	Name   string
	Sort   int
	Status int8
	Extend map[string]any
}

// ===== 仓库层预定义错误 =====
var (
	ErrDictNotFoundDB    = errors.New("dict not found in db")
	ErrDictItemNotFoundDB = errors.New("dict item not found in db")
)
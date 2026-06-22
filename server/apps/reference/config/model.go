// Package config 通用配置 - 领域模型与仓储接口
//
// 与 apps/reference/dict 对齐的设计点：
//   - scope ('platform' | 'tenant')：平台级 vs 租户级
//   - visibility ('all' | 'whitelist' | 'blacklist')：平台 group 对租户的可见性
//   - config_visibility 表：平台 group 对单个租户的访问级别
//   - platform_item_id + is_override：租户覆盖平台 item 的机制
//
// 与 dict 不同的语义：
//   - config_items 用 (category_id, key) 而不是 (dict_id, code) —— 因为配置以 key 而非 code 寻址
//   - 没有 scope=tenant 的"code 全局唯一"约束，tenant scope 也按 (tenant_id, code) 唯一
package config

import (
	"context"
	"time"
)

// ConfigCategory 配置分组
//
// scope='platform' 的行 tenant_id=0，由 super_admin 维护，
// 租户可通过 config_visibility 控制自己的可见性。
type ConfigCategory struct {
	ID         uint                   `json:"id"`
	TenantID   uint                   `json:"tenant_id"`
	Code       string                 `json:"code"`
	Name       string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Icon       *string                `json:"icon,omitempty"`
	Sort       int                    `json:"sort"`
	Scope      string                 `json:"scope"`      // 'platform' | 'tenant'
	Visibility string                 `json:"visibility"` // 'all' | 'whitelist' | 'blacklist'
	IsSystem   bool                   `json:"is_system"`
	IsPublic   bool                   `json:"is_public"`
	Status     int8                   `json:"status"`
	Extend     map[string]interface{} `json:"extend,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ConfigItem 配置项 = 定义 + 值
//
// platform_item_id 非空 + is_override=TRUE 表示这是租户对某 platform item 的覆盖。
// 同一 (tenant_id, platform_item_id) 仅一条 override（uk_config_item_override）。
type ConfigItem struct {
	ID             uint        `json:"id"`
	TenantID       uint        `json:"tenant_id"`
	CategoryID        uint        `json:"category_id"`
	Key            string      `json:"key"`
	Value          interface{} `json:"value,omitempty"`
	DefaultValue   interface{} `json:"default_value,omitempty"`
	Type           string      `json:"type"`
	Label          *string     `json:"label,omitempty"`
	Description    *string     `json:"description,omitempty"`
	Options        interface{} `json:"options,omitempty"`
	Validation     interface{} `json:"validation,omitempty"`
	Sort           int         `json:"sort"`
	IsPublic       bool        `json:"is_public"`
	IsReadonly     bool        `json:"is_readonly"`
	IsSystem       bool        `json:"is_system"`
	PlatformItemID *uint       `json:"platform_item_id,omitempty"`
	IsOverride     bool        `json:"is_override"`
	Status         int8        `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// ConfigVisibility 平台 group 对单个租户的访问级别
//
// 与 dict_visibility 同构，仅主键字段名不同（category_id vs dict_id）。
type ConfigVisibility struct {
	ID        uint      `json:"id"`
	CategoryID   uint      `json:"category_id"`
	TenantID  uint      `json:"tenant_id"`
	Access    string    `json:"access"` // 'invisible' | 'readonly' | 'editable'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryWithItems 分组 + 该分组下所有项（业务合并消费用）
type CategoryWithItems struct {
	Group ConfigCategory  `json:"group"`
	Items []ConfigItem `json:"items"`
}

// ResolvedConfig 业务最终消费的合并配置
//
// 平台 group + 租户 item override 的合并产物。
// 业务代码只需调 Resolver.Resolve() 即可拿到。
type ResolvedConfig struct {
	CategoryID     uint                    `json:"category_id"`
	CategoryCode   string                  `json:"group_code"`
	CategoryName   string                  `json:"group_name"`
	Items       map[string]ResolvedItem `json:"items"` // key -> item
}

// ResolvedItem 单项合并后的视图
//
// Source 标注 'platform' | 'override' | 'tenant'，便于前端调试。
type ResolvedItem struct {
	Key          string      `json:"key"`
	Value        interface{} `json:"value"`
	Type         string      `json:"type"`
	Label        *string     `json:"label,omitempty"`
	PlatformItemID *uint      `json:"platform_item_id,omitempty"`
	IsOverride   bool        `json:"is_override"`
	Source       string      `json:"source"` // 'platform' | 'override' | 'tenant'
}

// ConfigRepository 仓储接口（与 dict 对齐：Group + Item + Visibility 分层）
type ConfigRepository interface {
	// ============ Group ============
	ListGroups(ctx context.Context, tenantID uint) ([]ConfigCategory, error)
	ListPlatformGroups(ctx context.Context) ([]ConfigCategory, error)
	GetGroupByID(ctx context.Context, id uint) (*ConfigCategory, error)
	GetGroupByCode(ctx context.Context, tenantID uint, code string) (*ConfigCategory, error)
	GetPlatformGroupByCode(ctx context.Context, code string) (*ConfigCategory, error)
	CreateGroup(ctx context.Context, tenantID uint, scope string, req CreateGroupRepoReq) (*ConfigCategory, error)
	UpdateGroup(ctx context.Context, id uint, req UpdateGroupRepoReq) (*ConfigCategory, error)
	DeleteGroup(ctx context.Context, id uint) error

	// ============ Item ============
	ListItemsByGroup(ctx context.Context, categoryID uint) ([]ConfigItem, error)
	ListItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error)
	ListPlatformItemsByGroup(ctx context.Context, categoryID uint) ([]ConfigItem, error)
	ListPublicItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error)
	ListPublicItemsByGroupCode(ctx context.Context, tenantID uint, groupCode string) ([]ConfigItem, error)
	GetItemByID(ctx context.Context, id uint) (*ConfigItem, error)
	CreateItem(ctx context.Context, tenantID, categoryID uint, req CreateItemRepoReq) (*ConfigItem, error)
	UpdateItem(ctx context.Context, id uint, req UpdateItemRepoReq) (*ConfigItem, error)
	ResetItem(ctx context.Context, id uint) (*ConfigItem, error)
	DeleteItem(ctx context.Context, id uint) error
	CountItemsByGroup(ctx context.Context, categoryID uint) (int64, error)

	// ============ Override（租户覆盖平台 item）============
	UpsertOverride(ctx context.Context, tenantID, platformItemID uint, value interface{}) (*ConfigItem, error)
	DeleteOverride(ctx context.Context, tenantID, platformItemID uint) error

	// ============ Visibility（平台 group 对租户可见性）============
	ListVisibility(ctx context.Context, categoryID uint) ([]ConfigVisibility, error)
	UpsertVisibility(ctx context.Context, categoryID, tenantID uint, access string) (*ConfigVisibility, error)
	DeleteVisibility(ctx context.Context, categoryID, tenantID uint) error

	// ============ Resolve（业务合并消费）============
	ResolveGroupForTenant(ctx context.Context, tenantID uint, groupCode string) (*ResolvedConfig, error)
	ResolveAllForTenant(ctx context.Context, tenantID uint) (map[string]*ResolvedConfig, error)
}

// CreateGroupRepoReq 仓储层创建分组参数
type CreateGroupRepoReq struct {
	Code        string
	Name        string
	Description *string
	Icon        *string
	Sort        int
	IsSystem    bool
	IsPublic    bool
}

// UpdateGroupRepoReq 仓储层更新分组参数（部分字段可为 nil = 不更新）
type UpdateGroupRepoReq struct {
	Name        *string
	Description *string
	Icon        *string
	Sort        *int
	IsPublic    *bool
	Visibility  *string
	Status      *int8
}

// CreateItemRepoReq 仓储层创建项参数
type CreateItemRepoReq struct {
	Key          string
	Value        interface{}
	DefaultValue interface{}
	Type         string
	Label        *string
	Description  *string
	Options      interface{}
	Validation   interface{}
	Sort         int
	IsPublic     bool
	IsReadonly   bool
	IsSystem     bool
}

// UpdateItemRepoReq 仓储层更新项参数（部分字段可为 nil = 不更新）
type UpdateItemRepoReq struct {
	Value       *interface{}
	Label       *string
	Description *string
	Sort        *int
	IsPublic    *bool
	IsReadonly  *bool
	Status      *int8
}

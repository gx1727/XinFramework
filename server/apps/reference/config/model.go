// Package config 通用配置 - 领域模型与仓储接口
package config

import (
	"context"
	"time"
)

// ConfigGroup 配置分组
type ConfigGroup struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Sort        int       `json:"sort"`
	IsSystem    bool      `json:"is_system"`
	IsPublic    bool      `json:"is_public"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConfigItem 配置项 = 定义 + 值
type ConfigItem struct {
	ID           uint        `json:"id"`
	TenantID     uint        `json:"tenant_id"`
	GroupID      uint        `json:"group_id"`
	Key          string      `json:"key"`
	Value        interface{} `json:"value,omitempty"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Type         string      `json:"type"`
	Label        *string     `json:"label,omitempty"`
	Description  *string     `json:"description,omitempty"`
	Options      interface{} `json:"options,omitempty"`
	Validation   interface{} `json:"validation,omitempty"`
	Sort         int         `json:"sort"`
	IsPublic     bool        `json:"is_public"`
	IsReadonly   bool        `json:"is_readonly"`
	IsSystem     bool        `json:"is_system"`
	Status       int8        `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// GroupWithItems 分组 + 该分组下所有项
type GroupWithItems struct {
	Group ConfigGroup   `json:"group"`
	Items []ConfigItem  `json:"items"`
}

// ConfigRepository 仓储接口
type ConfigRepository interface {
	// Group
	ListGroups(ctx context.Context, tenantID uint) ([]ConfigGroup, error)
	GetGroupByID(ctx context.Context, id uint) (*ConfigGroup, error)
	GetGroupByCode(ctx context.Context, tenantID uint, code string) (*ConfigGroup, error)
	CreateGroup(ctx context.Context, tenantID uint, req CreateGroupRepoReq) (*ConfigGroup, error)
	UpdateGroup(ctx context.Context, id uint, req UpdateGroupRepoReq) (*ConfigGroup, error)
	DeleteGroup(ctx context.Context, id uint) error

	// Item
	ListItemsByGroup(ctx context.Context, groupID uint) ([]ConfigItem, error)
	ListItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error)
	ListPublicItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error)
	ListPublicItemsByGroupCode(ctx context.Context, tenantID uint, groupCode string) ([]ConfigItem, error)
	GetItemByID(ctx context.Context, id uint) (*ConfigItem, error)
	CreateItem(ctx context.Context, tenantID, groupID uint, req CreateItemRepoReq) (*ConfigItem, error)
	UpdateItem(ctx context.Context, id uint, req UpdateItemRepoReq) (*ConfigItem, error)
	ResetItem(ctx context.Context, id uint) (*ConfigItem, error)
	DeleteItem(ctx context.Context, id uint) error
	CountItemsByGroup(ctx context.Context, groupID uint) (int64, error)
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

// UpdateGroupRepoReq 仓储层更新分组参数
type UpdateGroupRepoReq struct {
	Name        *string
	Description *string
	Icon        *string
	Sort        *int
	IsPublic    *bool
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

// Package sysrole 实现"sys 域角色"管理 API（sys_roles 表）。
package sysrole

import (
	"context"
	"time"
)

// Role 是本模块的内部 Go struct，包装 sysauth.Role。
type Role struct {
	ID          uint      `json:"id"`
	OrgID       *uint     `json:"org_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DataScope   int8      `json:"data_scope"`
	IsDefault   bool      `json:"is_default"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuLite / PermissionLite 是关联 sys_menus / sys_permissions 的最小投影。
type MenuLite struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type PermissionLite struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	MenuID *uint  `json:"menu_id"`
}

type CreateRepoReq struct {
	OrgID       *uint
	Code        string
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
	CreatedBy   uint
}

type UpdateRepoReq struct {
	OrgID       *uint
	Code        *string
	Name        *string
	Description *string
	DataScope   *int8
	IsDefault   *bool
	Sort        *int
	Status      *int8
	UpdatedBy   uint
}

type Repository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, code string) (*Role, error)
	List(ctx context.Context, keyword string, page, size int) ([]Role, int64, error)
	Create(ctx context.Context, req CreateRepoReq) (*Role, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*Role, error)
	Delete(ctx context.Context, id uint, updatedBy uint) error

	ListUsers(ctx context.Context, roleID uint) ([]uint, error)
	ListMenus(ctx context.Context, roleID uint) ([]MenuLite, error)
	AssignMenus(ctx context.Context, roleID uint, menuIDs []uint) error
	ListPermissions(ctx context.Context, roleID uint) ([]PermissionLite, error)
	AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
}

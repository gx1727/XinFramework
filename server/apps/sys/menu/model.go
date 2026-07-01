// Package sysmenu 实现"sys 域菜单"管理 API（sys_menus 表）。
//
// 实现 sys_menus 表的 sys 菜单 API。原 apps/platform/menu 已迁入此处（Phase 0023.4）。
//
//	| 模块                  | 表                    | 用途                                              | 状态     |
//	| --------------------- | --------------------- | ------------------------------------------------- | -------- |
//	| (已删除)               | sys_menus             | sys 菜单唯一来源，物理独立 schema（Phase 0023+）| 终态        |
//	| apps/sys/menu         | sys_menus             | 新接口，对齐 tenant_menus 物理分离 schema（Phase 0023+）| 新建（终态）|
//
// 2026-06-23: 0023.4 完成，sys_menus 是 sys 菜单唯一来源。
package sysmenu

import (
	"context"
	"time"
)

type Menu struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Subtitle  *string   `json:"subtitle"`
	URL       *string   `json:"url"`
	Path      *string   `json:"path"`
	Icon      *string   `json:"icon"`
	Sort      int       `json:"sort"`
	ParentID  *uint     `json:"parent_id"`
	Ancestors *string   `json:"ancestors"`
	Visible   bool      `json:"visible"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Children  []*Menu   `json:"children,omitempty"`
}

type CreateRepoReq struct {
	Code      string
	Name      string
	Subtitle  *string
	URL       *string
	Path      *string
	Icon      *string
	Sort      int
	ParentID  *uint
	Ancestors *string
	Visible   bool
	Enabled   bool
	CreatedBy uint
}

type UpdateRepoReq struct {
	Code      *string
	Name      *string
	Subtitle  *string
	URL       *string
	Path      *string
	Icon      *string
	Sort      *int
	ParentID  *uint
	Ancestors *string
	Visible   *bool
	Enabled   *bool
	UpdatedBy uint
}

type Repository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, code string) (*Menu, error)
	GetAll(ctx context.Context) ([]Menu, error)
	// ListByUserRoles 返回被分配给 callerRoles 任一角色的菜单（多角色并集去重）。
	// 运行时接口"按角色过滤"专用，super_admin 仍走 GetAll。
	ListByUserRoles(ctx context.Context, accountID uint, callerRoles []string) ([]Menu, error)
	Create(ctx context.Context, req CreateRepoReq) (*Menu, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*Menu, error)
	Delete(ctx context.Context, id uint, updatedBy uint) error
}

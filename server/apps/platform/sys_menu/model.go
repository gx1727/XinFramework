// Package sysmenu 实现"平台域菜单"管理 API（sys_menus 表）。
//
// 重要：与 apps/platform/menu 是**两个不同模块**。
//
//	| 模块                  | 表                    | 用途                                              | 状态     |
//	| --------------------- | --------------------- | ------------------------------------------------- | -------- |
//	| apps/platform/menu    | menus (scope=platform)| 老接口，向 tenant_id=0 写入（混合 schema）         | 过渡期保留 |
//	| apps/platform/sys_menu| sys_menus             | 新接口，对齐 tenant_menus 物理分离 schema（Phase 0023+）| 新建（终态）|
//
// 终态：Phase 0023.4 切换后 apps/platform/menu 弃用。
// 过渡期：两个模块并存，不互通数据（mappers in Phase 0023.4 解决）。
package sysmenu

import (
	"context"
	"time"
)

type Menu struct {
	ID        uint       `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Subtitle  *string    `json:"subtitle"`
	URL       *string    `json:"url"`
	Path      *string    `json:"path"`
	Icon      *string    `json:"icon"`
	Sort      int        `json:"sort"`
	ParentID  *uint      `json:"parent_id"`
	Ancestors *string    `json:"ancestors"`
	Visible   bool       `json:"visible"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Children  []*Menu    `json:"children,omitempty"`
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
	Create(ctx context.Context, req CreateRepoReq) (*Menu, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*Menu, error)
	Delete(ctx context.Context, id uint, updatedBy uint) error
}

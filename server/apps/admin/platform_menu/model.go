// Package platformmenu 实现"平台级菜单"管理 API。
//
// 与 apps/rbac/menu 的关键区别：
//
//	| 维度          | apps/rbac/menu              | apps/admin/platform_menu      |
//	| tenant_id     | 来自 ctx.TenantID（租户内）  | 恒为 0（硬编码，平台级）       |
//	| DB 事务上下文 | RunInTenantTx                | RunInPlatformTx（bypass RLS）  |
//	| 中间件        | Require(permission.P(...))   | RequirePlatformRole("super_admin") |
//	| 路由前缀      | /api/v1/menus                | /api/v1/admin/platform-menus  |
//	| 错误码段      | 5001-5099                    | 15001-15999                   |
//
// 两套 API 共享同一张 menus 表（靠 tenant_id 区分），但代码完全独立，
// 避免反向依赖 apps/rbac/。
package platformmenu

import (
	"context"
	"time"
)

// Menu 与 rbac/menu.Menu 字段集完全一致（共享 menus 表 schema）。
//
// 注意：这是 type 定义，不是 alias —— 刻意保持独立，不让 rbac/menu
// 的字段变更牵动 platform_menu 的 SQL。
type Menu struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"` // 永远 = 0
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
}

// MenuRepository 是 platform_menu 的数据访问契约。
//
// 与 rbac/menu.MenuRepository 的区别：所有方法**不需要传 tenantID**
// —— 因为平台菜单固定在 tenant_id=0，参数化反而是泄漏点。
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, code string) (*Menu, error)
	GetAll(ctx context.Context) ([]Menu, error)
	Create(ctx context.Context, req CreateRepoReq) (*Menu, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*Menu, error)
	Delete(ctx context.Context, id uint) error
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
}

type UpdateRepoReq struct {
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
}

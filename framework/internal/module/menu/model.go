package menu

import (
	"context"
	"errors"
	"time"
)

// Menu represents a menu entity
type Menu struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Subtitle  string    `json:"subtitle"`
	URL       string    `json:"url"`
	Path      string    `json:"path"`
	Icon      string    `json:"icon"`
	Sort      int       `json:"sort"`
	ParentID  uint      `json:"parent_id"`
	Ancestors string    `json:"ancestors"`
	Visible   bool      `json:"visible"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MenuRepository defines data access operations for menus
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Menu, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Menu, error)
	GetUserMenus(ctx context.Context, tenantID, userID uint) ([]Menu, error)
	Create(ctx context.Context, tenantID uint, req CreateMenuRepoReq) (*Menu, error)
	Update(ctx context.Context, id uint, req UpdateMenuRepoReq) (*Menu, error)
	Delete(ctx context.Context, id uint) error
}

// CreateMenuRepoReq fields for menu creation
type CreateMenuRepoReq struct {
	Code      string
	Name      string
	Subtitle  string
	URL       string
	Path      string
	Icon      string
	Sort      int
	ParentID  uint
	Ancestors string
	Visible   bool
	Enabled   bool
}

// UpdateMenuRepoReq fields for menu update
type UpdateMenuRepoReq struct {
	Code     string
	Name     string
	Subtitle string
	URL      string
	Path     string
	Icon     string
	Sort     int
	Visible  bool
	Enabled  bool
}

var (
	ErrMenuNotFoundDB = errors.New("menu not found")
)

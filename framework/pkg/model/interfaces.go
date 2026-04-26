package model

import (
	"context"
	"time"
)

// ============ User Repository ============

// User represents a user entity
type User struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	AccountID uint      `json:"account_id"`
	Code      string    `json:"code"`
	Status    int8      `json:"status"`
	RealName  string    `json:"real_name"`
	Avatar    string    `json:"avatar"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository defines data access operations for users
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error)
	Create(ctx context.Context, tenantID, accountID uint, code string) (*User, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	Delete(ctx context.Context, id uint) error
}

// ============ Role Repository ============

// Role represents a role entity
type Role struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleRepository defines data access operations for roles
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Role, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	List(ctx context.Context, tenantID uint) ([]Role, error)
}

// ============ Account Repository ============

// Account represents a global account (cross-tenant)
type Account struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	RealName  string    `json:"real_name"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountRepository defines data access operations for accounts
type AccountRepository interface {
	GetByID(ctx context.Context, id uint) (*Account, error)
	GetByUsername(ctx context.Context, username string) (*Account, error)
	GetByPhone(ctx context.Context, phone string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*Account, error)
	Exists(ctx context.Context, account string) (bool, error)
}

// ============ Tenant Repository ============

// Tenant represents a tenant entity
type Tenant struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    int16     `json:"status"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Area      string    `json:"area"`
	Address   string    `json:"address"`
	Config    string    `json:"config"`
	Dashboard string    `json:"dashboard"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uint      `json:"created_by"`
	UpdatedBy uint      `json:"updated_by"`
	IsDeleted bool      `json:"is_deleted"`
}

// TenantRepository defines data access operations for tenants
type TenantRepository interface {
	GetByID(ctx context.Context, id uint) (*Tenant, error)
	GetByCode(ctx context.Context, code string) (*Tenant, error)
	List(ctx context.Context, keyword string, status *int16, page, size int) ([]Tenant, int64, error)
	Create(ctx context.Context, code, name, contact, phone, email string) (*Tenant, error)
	Update(ctx context.Context, id uint, name, contact, phone, email, province, city, area, address string) (*Tenant, error)
	Delete(ctx context.Context, id uint) error
}

// ============ Menu Repository ============

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
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MenuRepository defines data access operations for menus
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Menu, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Menu, error)
	GetUserMenus(ctx context.Context, tenantID, userID uint) ([]Menu, error)
}

// ============ Resource Repository ============

// Resource represents a resource/permission entity
type Resource struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	MenuID      uint      `json:"menu_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResourceRepository defines data access operations for resources
type ResourceRepository interface {
	GetByID(ctx context.Context, id uint) (*Resource, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Resource, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Resource, error)
	GetUserResources(ctx context.Context, tenantID, userID uint) ([]Resource, error)
}

// ============ CmsPost Repository ============

// CmsPost represents a CMS article entity
type CmsPost struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    int16     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// CmsPostRepository defines data access operations for CMS posts
type CmsPostRepository interface {
	GetByID(ctx context.Context, id uint) (*CmsPost, error)
	List(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]CmsPost, int64, error)
	Create(ctx context.Context, tenantID uint, title, content string, status int16) (*CmsPost, error)
	Update(ctx context.Context, id uint, title, content string, status int16) error
	Delete(ctx context.Context, id uint) error
}

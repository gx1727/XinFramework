package tenant

import (
	"context"
	"errors"
	"time"
)

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
	// CountActiveUsers 统计该租户下 is_deleted=FALSE 的用户数。
	// 删除前置校验：>0 时禁止软删租户，避免留下带活跃用户的幽灵租户。
	CountActiveUsers(ctx context.Context, tenantID uint) (int64, error)
	// UpdateStatus 仅修改 status 字段（如禁用 / 启用），与通用 Update 拆开便于审计与权限细分。
	UpdateStatus(ctx context.Context, id uint, status int16) (*Tenant, error)
}

var (
	ErrTenantNotFoundDB   = errors.New("tenant not found")
	ErrTenantCodeExistsDB = errors.New("tenant code already exists")
)

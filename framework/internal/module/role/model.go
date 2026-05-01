package role

import (
	"context"
	"errors"
	"time"
)

// Role represents a role entity
type Role struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	OrgID       uint      `json:"org_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DataScope   int8      `json:"data_scope"`
	Extend      string    `json:"extend"`
	IsDefault   bool      `json:"is_default"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleRepository defines data access operations for roles
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Role, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Role, int64, error)
	Create(ctx context.Context, tenantID uint, req CreateRoleRepoReq) (*Role, error)
	Update(ctx context.Context, id uint, req UpdateRoleRepoReq) (*Role, error)
	Delete(ctx context.Context, id uint) error
}

// CreateRoleRepoReq fields for role creation
type CreateRoleRepoReq struct {
	Code        string
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
}

// UpdateRoleRepoReq fields for role update
type UpdateRoleRepoReq struct {
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
}

var (
	ErrRoleNotFoundDB        = errors.New("role not found")
	ErrDefaultRoleNotFoundDB = errors.New("default role not found")
)

package rbac

import (
	"context"
	"time"
)

// Role is the cross-module role representation.
type Role struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	OrgID       *uint     `json:"org_id"`
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

// RoleRepository is the cross-module role data access contract.
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Role, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Role, int64, error)
}

// globalRoleFactory is set by apps/rbac/role's init().
var globalRoleFactory func() RoleRepository

// Register wires a RoleRepository factory.
func RegisterRoleRepository(f func() RoleRepository) {
	globalRoleFactory = f
}

// GetRoleRepository returns the registered factory, or nil.
func GetRoleRepository() func() RoleRepository {
	return globalRoleFactory
}
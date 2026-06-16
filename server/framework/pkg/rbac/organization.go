package rbac

import (
	"context"
	"time"
)

// Organization is the cross-module org representation.
type Organization struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	ParentID    uint      `json:"parent_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AdminCode   string    `json:"admin_code"`
	Ancestors   string    `json:"ancestors"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrganizationRepository is the cross-module org data access contract.
type OrganizationRepository interface {
	GetByID(ctx context.Context, id uint) (*Organization, error)
	GetByIDScoped(ctx context.Context, id uint) (*Organization, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Organization, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Organization, error)
	GetTree(ctx context.Context, tenantID uint) ([]Organization, error)
	CountUsersInOrgTree(ctx context.Context, orgID uint) (int64, error)
}

// globalOrganizationFactory is set by apps/rbac/organization's init().
var globalOrganizationFactory func() OrganizationRepository

// Register wires an OrganizationRepository factory.
func RegisterOrganizationRepository(f func() OrganizationRepository) {
	globalOrganizationFactory = f
}

// GetOrganizationRepository returns the registered factory, or nil.
func GetOrganizationRepository() func() OrganizationRepository {
	return globalOrganizationFactory
}
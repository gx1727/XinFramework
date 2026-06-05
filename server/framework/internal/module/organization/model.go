package organization

import (
	"context"
	"time"
)

// Organization represents an organization entity
type Organization struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AdminCode   string    `json:"admin_code"`
	ParentID    uint      `json:"parent_id"`
	Ancestors   string    `json:"ancestors"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrganizationRepository defines data access operations for organizations
type OrganizationRepository interface {
	GetByID(ctx context.Context, id uint) (*Organization, error)
	GetByIDScoped(ctx context.Context, id uint) (*Organization, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Organization, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Organization, error)
	GetByTenantScoped(ctx context.Context, tenantID uint) ([]Organization, error)
	GetChildren(ctx context.Context, parentID uint) ([]Organization, error)
	GetChildrenScoped(ctx context.Context, parentID uint) ([]Organization, error)
	CountChildren(ctx context.Context, parentID uint) (int64, error)

	// CountUsersInOrgTree 统计直接挂在 orgID 下、或挂在 orgID 任意后代下的未删用户数。
	// 这里直接查 users 表（org 模块读 user 表是允许的，反过来则不行）。
	CountUsersInOrgTree(ctx context.Context, orgID uint) (int64, error)
	GetTree(ctx context.Context, tenantID uint) ([]Organization, error)
	GetTreeScoped(ctx context.Context, tenantID uint) ([]Organization, error)
	Create(ctx context.Context, tenantID uint, req CreateOrgRepoReq) (*Organization, error)
	Update(ctx context.Context, id uint, req UpdateOrgRepoReq) (*Organization, error)
	Delete(ctx context.Context, id uint) error
}

// CreateOrgRepoReq fields for organization creation
type CreateOrgRepoReq struct {
	Code        string
	Name        string
	Type        string
	Description string
	AdminCode   string
	ParentID    uint
	Ancestors   string
	Sort        int
	Status      int8
}

// UpdateOrgRepoReq fields for organization update
type UpdateOrgRepoReq struct {
	Name        string
	Type        string
	Description string
	AdminCode   string
	Sort        int
	Status      int8
}

package resource

import (
	"context"
	"errors"
	"time"
)

// Resource represents a resource/permission entity
type Resource struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	MenuID      uint      `json:"menu_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResourceRepository defines data access operations for resources
type ResourceRepository interface {
	GetByID(ctx context.Context, id uint) (*Resource, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Resource, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Resource, error)
	GetByMenu(ctx context.Context, menuID uint) ([]Resource, error)
	GetUserResources(ctx context.Context, tenantID, userID uint) ([]Resource, error)
	Create(ctx context.Context, tenantID uint, req CreateResourceRepoReq) (*Resource, error)
	Update(ctx context.Context, id uint, req UpdateResourceRepoReq) (*Resource, error)
	Delete(ctx context.Context, id uint) error
}

// CreateResourceRepoReq fields for resource creation
type CreateResourceRepoReq struct {
	MenuID      uint
	Code        string
	Name        string
	Action      string
	Description string
	Sort        int
	Status      int8
}

// UpdateResourceRepoReq fields for resource update
type UpdateResourceRepoReq struct {
	Name        string
	Action      string
	Description string
	Sort        int
	Status      int8
}

var (
	ErrResourceNotFoundDB = errors.New("resource not found")
)

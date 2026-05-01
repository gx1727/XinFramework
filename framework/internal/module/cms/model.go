package cms

import (
	"context"
	"time"
)

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

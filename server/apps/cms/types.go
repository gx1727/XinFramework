package cms

import (
	"context"
	"time"

	"gx1727.com/xin/framework/pkg/extapi"
)

// CmsPost CMS 文章模型
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

// User 用户模型（别名，通过 extapi 获取）
type User = extapi.User

// Tenant 租户模型（别名，通过 extapi 获取）
type Tenant = extapi.Tenant

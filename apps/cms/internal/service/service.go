package service

import (
	"context"
	"fmt"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/model"
)

type Service struct {
	userRepo   model.UserRepository
	tenantRepo model.TenantRepository
	// postRepo 不通过依赖注入，直接使用 db.Get() 在方法中访问
}

// NewService 创建 Service 实例，只注入 Framework 的 Repository
func NewService(
	userRepo model.UserRepository,
	tenantRepo model.TenantRepository,
) *Service {
	return &Service{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

func (s *Service) GetUser(ctx context.Context, userID uint) (*model.User, error) {
	if s.userRepo == nil {
		return &model.User{ID: userID, Code: "demo", RealName: "Demo User"}, nil
	}
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) ListUsers(ctx context.Context, tenantID uint, keyword string, page, size int) ([]model.User, int64, error) {
	if s.userRepo == nil {
		return []model.User{}, 0, nil
	}
	return s.userRepo.List(ctx, tenantID, keyword, page, size)
}

func (s *Service) GetTenant(ctx context.Context, tenantID uint) (*model.Tenant, error) {
	if s.tenantRepo == nil {
		return &model.Tenant{ID: tenantID, Code: "demo", Name: "Demo Tenant"}, nil
	}
	return s.tenantRepo.GetByID(ctx, tenantID)
}

func (s *Service) ListPosts(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]model.CmsPost, int64, error) {
	pool := db.Get()
	if pool == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	// TODO: 实现 CMS Post 的查询逻辑
	// 这里应该创建 CmsPostRepository 或直接使用 SQL 查询
	return []model.CmsPost{}, 0, nil
}

func (s *Service) GetPost(ctx context.Context, id uint) (*model.CmsPost, error) {
	pool := db.Get()
	if pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// TODO: 实现 CMS Post 的查询逻辑
	var post model.CmsPost
	err := pool.QueryRow(ctx,
		"SELECT id, tenant_id, title, content, status, created_at, updated_at FROM cms_posts WHERE id = $1 AND is_deleted = false",
		id,
	).Scan(&post.ID, &post.TenantID, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	return &post, nil
}

func (s *Service) CreatePost(ctx context.Context, tenantID uint, title, content string, status int16) (*model.CmsPost, error) {
	pool := db.Get()
	if pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var post model.CmsPost
	err := pool.QueryRow(ctx,
		"INSERT INTO cms_posts (tenant_id, title, content, status) VALUES ($1, $2, $3, $4) RETURNING id, tenant_id, title, content, status, created_at, updated_at",
		tenantID, title, content, status,
	).Scan(&post.ID, &post.TenantID, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return &post, nil
}

func (s *Service) UpdatePost(ctx context.Context, id uint, title, content string, status int16) error {
	pool := db.Get()
	if pool == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := pool.Exec(ctx,
		"UPDATE cms_posts SET title = $1, content = $2, status = $3, updated_at = NOW() WHERE id = $4 AND is_deleted = false",
		title, content, status, id,
	)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	return nil
}

func (s *Service) DeletePost(ctx context.Context, id uint) error {
	pool := db.Get()
	if pool == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := pool.Exec(ctx,
		"UPDATE cms_posts SET is_deleted = true, updated_at = NOW() WHERE id = $1",
		id,
	)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

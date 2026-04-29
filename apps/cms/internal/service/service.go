package service

import (
	"context"
	"fmt"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/model"
)

type Service struct{}

// NewService 创建 Service 实例
func NewService() *Service {
	return &Service{}
}

func (s *Service) GetUser(ctx context.Context, userID uint) (*model.User, error) {
	pool := db.Get()
	if pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var user model.User
	err := pool.QueryRow(ctx,
		"SELECT id, tenant_id, account_id, code, real_name, email, phone, status, created_at, updated_at FROM users WHERE id = $1 AND is_deleted = false",
		userID,
	).Scan(&user.ID, &user.TenantID, &user.AccountID, &user.Code, &user.RealName, &user.Email, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

func (s *Service) ListUsers(ctx context.Context, tenantID uint, keyword string, page, size int) ([]model.User, int64, error) {
	pool := db.Get()
	if pool == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	offset := (page - 1) * size

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND is_deleted = false"
	countArgs := []interface{}{tenantID}

	if keyword != "" {
		countQuery += " AND (real_name LIKE $2 OR code LIKE $2 OR email LIKE $2)"
		countArgs = append(countArgs, "%"+keyword+"%")
	}

	err := pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// 查询列表
	query := "SELECT id, tenant_id, account_id, code, real_name, email, phone, status, created_at, updated_at FROM users WHERE tenant_id = $1 AND is_deleted = false"
	args := []interface{}{tenantID}

	if keyword != "" {
		query += " AND (real_name LIKE $2 OR code LIKE $2 OR email LIKE $2)"
		args = append(args, "%"+keyword+"%")
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, size, offset)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.TenantID, &user.AccountID, &user.Code, &user.RealName, &user.Email, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

func (s *Service) GetTenant(ctx context.Context, tenantID uint) (*model.Tenant, error) {
	pool := db.Get()
	if pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var tenant model.Tenant
	err := pool.QueryRow(ctx,
		"SELECT id, code, name, status, contact, phone, email, province, city, area, address, config, dashboard, created_at, updated_at, created_by, updated_by FROM tenants WHERE id = $1 AND is_deleted = false",
		tenantID,
	).Scan(&tenant.ID, &tenant.Code, &tenant.Name, &tenant.Status, &tenant.Contact, &tenant.Phone, &tenant.Email, &tenant.Province, &tenant.City, &tenant.Area, &tenant.Address, &tenant.Config, &tenant.Dashboard, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.CreatedBy, &tenant.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return &tenant, nil
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

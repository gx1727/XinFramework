package cms

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository 数据访问层
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository 创建 Repository 实例
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ============ User 查询 ============

func (r *Repository) GetUserByID(ctx context.Context, userID uint) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, code, real_name, email, phone, status, created_at, updated_at 
		 FROM users WHERE id = $1 AND is_deleted = false`,
		userID,
	).Scan(&user.ID, &user.TenantID, &user.Code, &user.RealName, &user.Email, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

func (r *Repository) ListUsers(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error) {
	offset := (page - 1) * size

	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND is_deleted = false`
	countArgs := []interface{}{tenantID}

	if keyword != "" {
		countQuery += ` AND (real_name LIKE $2 OR code LIKE $2 OR email LIKE $2)`
		countArgs = append(countArgs, "%"+keyword+"%")
	}

	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// 查询列表
	query := `SELECT id, tenant_id, code, real_name, email, phone, status, created_at, updated_at 
			  FROM users WHERE tenant_id = $1 AND is_deleted = false`
	args := []interface{}{tenantID}

	if keyword != "" {
		query += ` AND (real_name LIKE $2 OR code LIKE $2 OR email LIKE $2)`
		args = append(args, "%"+keyword+"%")
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, size, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.TenantID, &user.Code, &user.RealName, &user.Email, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// ============ Tenant 查询 ============

func (r *Repository) GetTenantByID(ctx context.Context, tenantID uint) (*Tenant, error) {
	var tenant Tenant
	err := r.db.QueryRow(ctx,
		`SELECT id, code, name, status, contact, phone, email, province, city, area, address, config, dashboard, 
				created_at, updated_at, created_by, updated_by 
		 FROM tenants WHERE id = $1 AND is_deleted = false`,
		tenantID,
	).Scan(&tenant.ID, &tenant.Code, &tenant.Name, &tenant.Status, &tenant.Contact, &tenant.Phone, &tenant.Email,
		&tenant.Province, &tenant.City, &tenant.Area, &tenant.Address, &tenant.Config, &tenant.Dashboard,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.CreatedBy, &tenant.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return &tenant, nil
}

// ============ CmsPost CRUD ============

func (r *Repository) ListPosts(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]CmsPost, int64, error) {
	offset := (page - 1) * size

	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM cms_posts WHERE tenant_id = $1 AND is_deleted = false`
	countArgs := []interface{}{tenantID}

	if keyword != "" {
		countQuery += ` AND (title LIKE $2 OR content LIKE $2)`
		countArgs = append(countArgs, "%"+keyword+"%")
	}

	if status != nil {
		countQuery += fmt.Sprintf(` AND status = $%d`, len(countArgs)+1)
		countArgs = append(countArgs, *status)
	}

	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	// 查询列表
	query := `SELECT id, tenant_id, title, content, status, created_at, updated_at 
			  FROM cms_posts WHERE tenant_id = $1 AND is_deleted = false`
	args := []interface{}{tenantID}

	if keyword != "" {
		query += ` AND (title LIKE $2 OR content LIKE $2)`
		args = append(args, "%"+keyword+"%")
	}

	if status != nil {
		query += fmt.Sprintf(` AND status = $%d`, len(args)+1)
		args = append(args, *status)
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, size, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []CmsPost
	for rows.Next() {
		var post CmsPost
		err := rows.Scan(&post.ID, &post.TenantID, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, total, nil
}

func (r *Repository) GetPostByID(ctx context.Context, id uint) (*CmsPost, error) {
	var post CmsPost
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, title, content, status, created_at, updated_at 
		 FROM cms_posts WHERE id = $1 AND is_deleted = false`,
		id,
	).Scan(&post.ID, &post.TenantID, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	return &post, nil
}

func (r *Repository) CreatePost(ctx context.Context, tenantID uint, title, content string, status int16) (*CmsPost, error) {
	var post CmsPost
	err := r.db.QueryRow(ctx,
		`INSERT INTO cms_posts (tenant_id, title, content, status) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, tenant_id, title, content, status, created_at, updated_at`,
		tenantID, title, content, status,
	).Scan(&post.ID, &post.TenantID, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return &post, nil
}

func (r *Repository) UpdatePost(ctx context.Context, id uint, title, content string, status int16) error {
	_, err := r.db.Exec(ctx,
		`UPDATE cms_posts SET title = $1, content = $2, status = $3, updated_at = NOW() 
		 WHERE id = $4 AND is_deleted = false`,
		title, content, status, id,
	)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	return nil
}

func (r *Repository) DeletePost(ctx context.Context, id uint) error {
	_, err := r.db.Exec(ctx,
		`UPDATE cms_posts SET is_deleted = true, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

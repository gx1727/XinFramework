package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/model"
)

type CmsPostRepository struct {
	db *pgxpool.Pool
}

func NewCmsPostRepository(db *pgxpool.Pool) model.CmsPostRepository {
	return &CmsPostRepository{db: db}
}

func (r *CmsPostRepository) GetByID(ctx context.Context, id uint) (*model.CmsPost, error) {
	var p model.CmsPost
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, content, status, created_at, updated_at, is_deleted
		FROM cms_posts
		WHERE id = $1 AND is_deleted = FALSE
	`, id).Scan(&p.ID, &p.TenantID, &p.Title, &p.Content, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("get cms post: %w", err)
	}
	return &p, nil
}

func (r *CmsPostRepository) List(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]model.CmsPost, int64, error) {
	where := "WHERE tenant_id = $1 AND is_deleted = FALSE"
	args := []interface{}{tenantID}
	argIdx := 2

	if keyword != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+keyword+"%")
		argIdx++
	}
	if status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}

	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM cms_posts "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count cms posts: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`
		SELECT id, tenant_id, title, content, status, created_at, updated_at, is_deleted
		FROM cms_posts %s
		ORDER BY id DESC LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list cms posts: %w", err)
	}
	defer rows.Close()

	var list []model.CmsPost
	for rows.Next() {
		var p model.CmsPost
		err := rows.Scan(&p.ID, &p.TenantID, &p.Title, &p.Content, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.IsDeleted)
		if err != nil {
			return nil, 0, fmt.Errorf("scan cms post: %w", err)
		}
		list = append(list, p)
	}

	return list, total, nil
}

func (r *CmsPostRepository) Create(ctx context.Context, tenantID uint, title, content string, status int16) (*model.CmsPost, error) {
	var p model.CmsPost
	err := r.db.QueryRow(ctx, `
		INSERT INTO cms_posts (tenant_id, title, content, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tenant_id, title, content, status, created_at, updated_at, is_deleted
	`, tenantID, title, content, status).Scan(&p.ID, &p.TenantID, &p.Title, &p.Content, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("create cms post: %w", err)
	}
	return &p, nil
}

func (r *CmsPostRepository) Update(ctx context.Context, id uint, title, content string, status int16) error {
	_, err := r.db.Exec(ctx, `
		UPDATE cms_posts
		SET title = $2, content = $3, status = $4, updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE
	`, id, title, content, status)
	if err != nil {
		return fmt.Errorf("update cms post: %w", err)
	}
	return nil
}

func (r *CmsPostRepository) Delete(ctx context.Context, id uint) error {
	_, err := r.db.Exec(ctx, `
		UPDATE cms_posts SET is_deleted = TRUE, updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete cms post: %w", err)
	}
	return nil
}

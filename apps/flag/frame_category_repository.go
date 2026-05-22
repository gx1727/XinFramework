package flag

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// FrameCategoryRepository 相框分类数据访问层
type FrameCategoryRepository struct {
	db *pgxpool.Pool
}

func NewFrameCategoryRepository(pool *pgxpool.Pool) *FrameCategoryRepository {
	return &FrameCategoryRepository{db: pool}
}

func (r *FrameCategoryRepository) List(ctx context.Context) (_ []FrameCategory, err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, type, sort, status
		FROM flag_frame_categories
		WHERE is_deleted = FALSE
		ORDER BY sort ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []FrameCategory
	for rows.Next() {
		var c FrameCategory
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &c.Type, &c.Sort, &c.Status); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *FrameCategoryRepository) GetByID(ctx context.Context, id uint) (_ *FrameCategory, err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	var c FrameCategory
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, type, sort, status
		FROM flag_frame_categories
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&c.ID, &c.TenantID, &c.Code, &c.Name, &c.Type, &c.Sort, &c.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFrameNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *FrameCategoryRepository) Create(ctx context.Context, c *FrameCategory) (_ *FrameCategory, err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	var result FrameCategory
	err = q.QueryRow(ctx, `
		INSERT INTO flag_frame_categories (tenant_id, code, name, type, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, code, name, type, sort, status`,
		c.TenantID, c.Code, c.Name, c.Type, c.Sort, c.Status).Scan(
		&result.ID, &result.TenantID, &result.Code, &result.Name, &result.Type, &result.Sort, &result.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrCategoryCodeExists.WithMsg(fmt.Sprintf("相框分类编码已存在: %v", err))
		}
		return nil, ErrCreateCategoryFailed.WithMsg(fmt.Sprintf("创建相框分类失败: %v", err))
	}
	return &result, nil
}

func (r *FrameCategoryRepository) Update(ctx context.Context, c *FrameCategory) (err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE flag_frame_categories SET code = $2, name = $3, type = $4, sort = $5, status = $6, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`,
		c.ID, c.Code, c.Name, c.Type, c.Sort, c.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrCategoryCodeExists.WithMsg(fmt.Sprintf("相框分类编码已存在: %v", err))
		}
		return ErrUpdateCategoryFailed.WithMsg(fmt.Sprintf("更新相框分类失败: %v", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

func (r *FrameCategoryRepository) Delete(ctx context.Context, id uint) (err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE flag_frame_categories SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete frame category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

package flag

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

// AvatarCategoryRepository 头像分类数据访问层
type AvatarCategoryRepository struct {
	db *pgxpool.Pool
}

func NewAvatarCategoryRepository(pool *pgxpool.Pool) *AvatarCategoryRepository {
	return &AvatarCategoryRepository{db: pool}
}

func (r *AvatarCategoryRepository) List(ctx context.Context) (_ []AvatarCategory, err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, code, name, icon, type, sort, status
		FROM flag_avatar_categories
		WHERE is_deleted = FALSE
		ORDER BY sort ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AvatarCategory
	for rows.Next() {
		var c AvatarCategory
		var icon *string
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &icon, &c.Type, &c.Sort, &c.Status); err != nil {
			return nil, err
		}
		if icon != nil {
			c.Icon = *icon
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *AvatarCategoryRepository) Create(ctx context.Context, c *AvatarCategory) (_ *AvatarCategory, err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}

	var result AvatarCategory
	var icon *string
	err = q.QueryRow(ctx, `
		INSERT INTO flag_avatar_categories (tenant_id, code, name, icon, type, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, code, name, icon, type, sort, status`,
		c.TenantID, c.Code, c.Name, nullStr(c.Icon), c.Type, c.Sort, c.Status).Scan(
		&result.ID, &result.TenantID, &result.Code, &result.Name, &icon, &result.Type, &result.Sort, &result.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrCategoryCodeExists.WithMsg(fmt.Sprintf("头像分类编码已存在: %v", err))
		}
		return nil, ErrCreateCategoryFailed.WithMsg(fmt.Sprintf("创建头像分类失败: %v", err))
	}
	if icon != nil {
		result.Icon = *icon
	}
	return &result, nil
}

func (r *AvatarCategoryRepository) Update(ctx context.Context, c *AvatarCategory) (err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE flag_avatar_categories SET code = $2, name = $3, icon = $4, type = $5, sort = $6, status = $7, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`,
		c.ID, c.Code, c.Name, nullStr(c.Icon), c.Type, c.Sort, c.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrCategoryCodeExists.WithMsg(fmt.Sprintf("头像分类编码已存在: %v", err))
		}
		return ErrUpdateCategoryFailed.WithMsg(fmt.Sprintf("更新头像分类失败: %v", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

func (r *AvatarCategoryRepository) Delete(ctx context.Context, id uint) (err error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, "SELECT set_config('app.show_deleted', $1, true)", "true")
	if err != nil {
		return err
	}

	tag, err := q.Exec(ctx, `
		UPDATE flag_avatar_categories SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete avatar category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

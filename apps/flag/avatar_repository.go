package flag

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
)

// AvatarRepository 头像数据访问层
type AvatarRepository struct {
	db *pgxpool.Pool
}

func NewAvatarRepository(pool *pgxpool.Pool) *AvatarRepository {
	return &AvatarRepository{db: pool}
}

func (r *AvatarRepository) List(ctx context.Context, categoryID, userID uint, avatarType string, page, size int) ([]Avatar, int64, error) {
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer release()

	if conn, ok := q.(*db.Conn); ok {
		if tid, ok := xincontext.TenantIDFrom(ctx); ok {
			_ = conn.SetTenant(ctx, tid)
		}
	}

	where := "WHERE is_deleted = FALSE"
	args := []interface{}{}
	argIdx := 1

	if categoryID > 0 {
		where += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, categoryID)
		argIdx++
	}
	if userID > 0 {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}
	if avatarType != "" {
		where += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, avatarType)
		argIdx++
	}

	var total int64
	if err := q.QueryRow(ctx, "SELECT COUNT(*) FROM flag_avatars "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	querySQL := fmt.Sprintf(`SELECT id, tenant_id, user_id, category_id, name, source_url, thumbnail_url, file_size, width, height, type, is_public, like_count, view_count, status, created_at, updated_at
		FROM flag_avatars %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := q.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Avatar
	for rows.Next() {
		var a Avatar
		var name, sourceURL, thumbnailURL *string
		var fileSize *int64
		var width, height *int
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.UserID, &a.CategoryID, &name, &sourceURL, &thumbnailURL,
			&fileSize, &width, &height, &a.Type, &a.IsPublic, &a.LikeCount, &a.ViewCount,
			&a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if name != nil {
			a.Name = *name
		}
		if sourceURL != nil {
			a.SourceURL = *sourceURL
		}
		if thumbnailURL != nil {
			a.ThumbnailURL = *thumbnailURL
		}
		if fileSize != nil {
			a.FileSize = *fileSize
		}
		if width != nil {
			a.Width = *width
		}
		if height != nil {
			a.Height = *height
		}
		list = append(list, a)
	}
	return list, total, nil
}

func (r *AvatarRepository) GetByID(ctx context.Context, id uint) (*Avatar, error) {
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	if conn, ok := q.(*db.Conn); ok {
		if tid, ok := xincontext.TenantIDFrom(ctx); ok {
			_ = conn.SetTenant(ctx, tid)
		}
	}

	var a Avatar
	var name, sourceURL, thumbnailURL *string
	var fileSize *int64
	var width, height *int
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, category_id, name, source_url, thumbnail_url, file_size, width, height, type, is_public, like_count, view_count, status, created_at, updated_at
		FROM flag_avatars
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&a.ID, &a.TenantID, &a.UserID, &a.CategoryID, &name, &sourceURL, &thumbnailURL,
		&fileSize, &width, &height, &a.Type, &a.IsPublic, &a.LikeCount, &a.ViewCount,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAvatarNotFound
		}
		return nil, err
	}
	if name != nil {
		a.Name = *name
	}
	if sourceURL != nil {
		a.SourceURL = *sourceURL
	}
	if thumbnailURL != nil {
		a.ThumbnailURL = *thumbnailURL
	}
	if fileSize != nil {
		a.FileSize = *fileSize
	}
	if width != nil {
		a.Width = *width
	}
	if height != nil {
		a.Height = *height
	}
	return &a, nil
}

func (r *AvatarRepository) Create(ctx context.Context, avatar *Avatar) (*Avatar, error) {
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	if conn, ok := q.(*db.Conn); ok {
		if tid, ok := xincontext.TenantIDFrom(ctx); ok {
			_ = conn.SetTenant(ctx, tid)
		}
	}

	var a Avatar
	var name, sourceURL, thumbnailURL *string
	var fileSize *int64
	var width, height *int
	err = q.QueryRow(ctx, `
		INSERT INTO flag_avatars (tenant_id, user_id, category_id, name, source_url, thumbnail_url, file_size, width, height, type, is_public, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, tenant_id, user_id, category_id, name, source_url, thumbnail_url, file_size, width, height, type, is_public, like_count, view_count, status, created_at, updated_at`,
		avatar.TenantID, avatar.UserID, nilIfZero(avatar.CategoryID), nullStr(avatar.Name),
		nullStr(avatar.SourceURL), nullStr(avatar.ThumbnailURL), avatar.FileSize,
		avatar.Width, avatar.Height, avatar.Type, avatar.IsPublic, avatar.Status).Scan(
		&a.ID, &a.TenantID, &a.UserID, &a.CategoryID, &name, &sourceURL, &thumbnailURL,
		&fileSize, &width, &height, &a.Type, &a.IsPublic, &a.LikeCount, &a.ViewCount,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create flag avatar: %w", err)
	}
	if name != nil {
		a.Name = *name
	}
	if sourceURL != nil {
		a.SourceURL = *sourceURL
	}
	if thumbnailURL != nil {
		a.ThumbnailURL = *thumbnailURL
	}
	if fileSize != nil {
		a.FileSize = *fileSize
	}
	if width != nil {
		a.Width = *width
	}
	if height != nil {
		a.Height = *height
	}
	return &a, nil
}

func (r *AvatarRepository) Update(ctx context.Context, avatar *Avatar) error {
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	defer release()

	if conn, ok := q.(*db.Conn); ok {
		if tid, ok := xincontext.TenantIDFrom(ctx); ok {
			_ = conn.SetTenant(ctx, tid)
		}
	}

	tag, err := q.Exec(ctx, `
		UPDATE flag_avatars SET
			name = $2, category_id = $3, is_public = $4, status = $5, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`,
		avatar.ID, nullStr(avatar.Name), nilIfZero(avatar.CategoryID), avatar.IsPublic, avatar.Status)
	if err != nil {
		return fmt.Errorf("update flag avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAvatarNotFound
	}
	return nil
}

func (r *AvatarRepository) Delete(ctx context.Context, id uint) error {
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return err
	}
	defer release()

	if conn, ok := q.(*db.Conn); ok {
		if tid, ok := xincontext.TenantIDFrom(ctx); ok {
			_ = conn.SetTenant(ctx, tid)
		}
	}

	tag, err := q.Exec(ctx, `
		UPDATE flag_avatars SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete flag avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAvatarNotFound
	}
	return nil
}

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

type FrameRepository struct {
	db *pgxpool.Pool
}

func NewFrameRepository(pool *pgxpool.Pool) *FrameRepository {
	return &FrameRepository{db: pool}
}

func (r *FrameRepository) List(ctx context.Context, categoryID uint, page, size int) ([]Frame, int64, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	where := "WHERE is_deleted = FALSE"
	args := []interface{}{}
	argIdx := 1

	if categoryID > 0 {
		where += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, categoryID)
		argIdx++
	}

	var total int64
	if err := conn.QueryRow(ctx, "SELECT COUNT(*) FROM flag_frames "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	querySQL := fmt.Sprintf(`SELECT id, tenant_id, category_id, name, description, preview_url, template_url, template_config, type, sort, status, created_at, updated_at
		FROM flag_frames %s ORDER BY sort ASC, id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := conn.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Frame
	for rows.Next() {
		var f Frame
		var description, previewURL, templateURL, templateConfig *string
		if err := rows.Scan(
			&f.ID, &f.TenantID, &f.CategoryID, &f.Name, &description, &previewURL,
			&templateURL, &templateConfig, &f.Type, &f.Sort, &f.Status,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if description != nil {
			f.Description = *description
		}
		if previewURL != nil {
			f.PreviewURL = *previewURL
		}
		if templateURL != nil {
			f.TemplateURL = *templateURL
		}
		if templateConfig != nil {
			f.TemplateConfig = *templateConfig
		}
		list = append(list, f)
	}
	return list, total, nil
}

func (r *FrameRepository) GetByID(ctx context.Context, id uint) (*Frame, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var f Frame
	var description, previewURL, templateURL, templateConfig *string
	err = conn.QueryRow(ctx, `
		SELECT id, tenant_id, category_id, name, description, preview_url, template_url, template_config, type, sort, status, created_at, updated_at
		FROM flag_frames
		WHERE is_deleted = FALSE AND id = $1`, id).Scan(
		&f.ID, &f.TenantID, &f.CategoryID, &f.Name, &description, &previewURL,
		&templateURL, &templateConfig, &f.Type, &f.Sort, &f.Status,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFrameNotFound
		}
		return nil, err
	}
	if description != nil {
		f.Description = *description
	}
	if previewURL != nil {
		f.PreviewURL = *previewURL
	}
	if templateURL != nil {
		f.TemplateURL = *templateURL
	}
	if templateConfig != nil {
		f.TemplateConfig = *templateConfig
	}
	return &f, nil
}

func (r *FrameRepository) Create(ctx context.Context, frame *Frame) (*Frame, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var f Frame
	var description, previewURL, templateURL, templateConfig *string
	err = conn.QueryRow(ctx, `
		INSERT INTO flag_frames (tenant_id, category_id, name, description, preview_url, template_url, template_config, type, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, tenant_id, category_id, name, description, preview_url, template_url, template_config, type, sort, status, created_at, updated_at`,
		frame.TenantID, nilIfZero(frame.CategoryID), frame.Name, nullStr(frame.Description),
		nullStr(frame.PreviewURL), nullStr(frame.TemplateURL), nullStr(frame.TemplateConfig),
		frame.Type, frame.Sort, frame.Status).Scan(
		&f.ID, &f.TenantID, &f.CategoryID, &f.Name, &description, &previewURL,
		&templateURL, &templateConfig, &f.Type, &f.Sort, &f.Status,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create flag frame: %w", err)
	}
	if description != nil {
		f.Description = *description
	}
	if previewURL != nil {
		f.PreviewURL = *previewURL
	}
	if templateURL != nil {
		f.TemplateURL = *templateURL
	}
	if templateConfig != nil {
		f.TemplateConfig = *templateConfig
	}
	return &f, nil
}

func (r *FrameRepository) Update(ctx context.Context, frame *Frame) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
		UPDATE flag_frames SET
			category_id = $2, name = $3, description = $4, preview_url = $5,
			template_url = $6, template_config = $7, type = $8, sort = $9, status = $10, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`,
		frame.ID, nilIfZero(frame.CategoryID), frame.Name, nullStr(frame.Description),
		nullStr(frame.PreviewURL), nullStr(frame.TemplateURL), nullStr(frame.TemplateConfig),
		frame.Type, frame.Sort, frame.Status)
	if err != nil {
		return fmt.Errorf("update flag frame: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

func (r *FrameRepository) Delete(ctx context.Context, id uint) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
		UPDATE flag_frames SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete flag frame: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

// ==================== Avatar Repository ====================

type AvatarRepository struct {
	db *pgxpool.Pool
}

func NewAvatarRepository(pool *pgxpool.Pool) *AvatarRepository {
	return &AvatarRepository{db: pool}
}

func (r *AvatarRepository) List(ctx context.Context, categoryID, userID uint, avatarType string, page, size int) ([]Avatar, int64, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
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
	if err := conn.QueryRow(ctx, "SELECT COUNT(*) FROM flag_avatars "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	querySQL := fmt.Sprintf(`SELECT id, tenant_id, user_id, category_id, name, source_url, thumbnail_url, file_size, width, height, type, is_public, like_count, view_count, status, created_at, updated_at
		FROM flag_avatars %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := conn.Query(ctx, querySQL, args...)
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
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var a Avatar
	var name, sourceURL, thumbnailURL *string
	var fileSize *int64
	var width, height *int
	err = conn.QueryRow(ctx, `
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
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var a Avatar
	var name, sourceURL, thumbnailURL *string
	var fileSize *int64
	var width, height *int
	err = conn.QueryRow(ctx, `
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
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
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
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
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

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nilIfZero(v uint) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

// ==================== Frame Category Repository ====================

type FrameCategoryRepository struct {
	db *pgxpool.Pool
}

func NewFrameCategoryRepository(pool *pgxpool.Pool) *FrameCategoryRepository {
	return &FrameCategoryRepository{db: pool}
}

func (r *FrameCategoryRepository) List(ctx context.Context) ([]FrameCategory, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	rows, err := conn.Query(ctx, `
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

func (r *FrameCategoryRepository) GetByID(ctx context.Context, id uint) (*FrameCategory, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var c FrameCategory
	err = conn.QueryRow(ctx, `
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

func (r *FrameCategoryRepository) Create(ctx context.Context, c *FrameCategory) (*FrameCategory, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var result FrameCategory
	err = conn.QueryRow(ctx, `
		INSERT INTO flag_frame_categories (tenant_id, code, name, type, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, code, name, type, sort, status`,
		c.TenantID, c.Code, c.Name, c.Type, c.Sort, c.Status).Scan(
		&result.ID, &result.TenantID, &result.Code, &result.Name, &result.Type, &result.Sort, &result.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("create frame category: %w", err)
	}
	return &result, nil
}

func (r *FrameCategoryRepository) Update(ctx context.Context, c *FrameCategory) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
		UPDATE flag_frame_categories SET code = $2, name = $3, type = $4, sort = $5, status = $6, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`,
		c.ID, c.Code, c.Name, c.Type, c.Sort, c.Status)
	if err != nil {
		return fmt.Errorf("update frame category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrFrameNotFound
	}
	return nil
}

func (r *FrameCategoryRepository) Delete(ctx context.Context, id uint) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
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

// ==================== Avatar Category Repository ====================

type AvatarCategoryRepository struct {
	db *pgxpool.Pool
}

func NewAvatarCategoryRepository(pool *pgxpool.Pool) *AvatarCategoryRepository {
	return &AvatarCategoryRepository{db: pool}
}

func (r *AvatarCategoryRepository) List(ctx context.Context, catType string) ([]AvatarCategory, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	where := "WHERE is_deleted = FALSE"
	args := []interface{}{}
	argIdx := 1
	if catType != "" {
		where += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, catType)
		argIdx++
	}

	querySQL := fmt.Sprintf(`SELECT id, tenant_id, code, name, icon, type, sort, status
		FROM flag_avatar_categories %s ORDER BY sort ASC, id ASC`, where)

	rows, err := conn.Query(ctx, querySQL, args...)
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

func (r *AvatarCategoryRepository) Create(ctx context.Context, c *AvatarCategory) (*AvatarCategory, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	var result AvatarCategory
	var icon *string
	err = conn.QueryRow(ctx, `
		INSERT INTO flag_avatar_categories (tenant_id, code, name, icon, type, sort, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, code, name, icon, type, sort, status`,
		c.TenantID, c.Code, c.Name, nullStr(c.Icon), c.Type, c.Sort, c.Status).Scan(
		&result.ID, &result.TenantID, &result.Code, &result.Name, &icon, &result.Type, &result.Sort, &result.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("create avatar category: %w", err)
	}
	if icon != nil {
		result.Icon = *icon
	}
	return &result, nil
}

func (r *AvatarCategoryRepository) Update(ctx context.Context, c *AvatarCategory) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
		UPDATE flag_avatar_categories SET code = $2, name = $3, icon = $4, type = $5, sort = $6, status = $7, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`,
		c.ID, c.Code, c.Name, nullStr(c.Icon), c.Type, c.Sort, c.Status)
	if err != nil {
		return fmt.Errorf("update avatar category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAvatarNotFound
	}
	return nil
}

func (r *AvatarCategoryRepository) Delete(ctx context.Context, id uint) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if tid, ok := xincontext.TenantIDFrom(ctx); ok {
		_ = conn.SetTenant(ctx, tid)
	}

	tag, err := conn.Exec(ctx, `
		UPDATE flag_avatar_categories SET is_deleted = TRUE, updated_at = NOW()
		WHERE is_deleted = FALSE AND id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete avatar category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAvatarNotFound
	}
	return nil
}

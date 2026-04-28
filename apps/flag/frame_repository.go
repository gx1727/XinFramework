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

// FrameRepository 相框数据访问层
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

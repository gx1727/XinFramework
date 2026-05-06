package asset

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/db"
)

type PostgresAttachmentRepository struct {
	db *pgxpool.Pool
}

func NewAttachmentRepository(db *pgxpool.Pool) *PostgresAttachmentRepository {
	return &PostgresAttachmentRepository{db: db}
}

func (r *PostgresAttachmentRepository) GetByID(ctx context.Context, id uint) (*Attachment, error) {
	q, release, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	query := `
		SELECT id, tenant_id, user_id, file_name, file_ext, mime_type, file_size, storage, object_key, url, hash, status, created_at, updated_at, is_deleted
		FROM attachments
		WHERE id = $1 AND is_deleted = false
	`
	var attachment Attachment
	var userID *uint
	err = q.QueryRow(ctx, query, id).Scan(
		&attachment.ID,
		&attachment.TenantID,
		&userID,
		&attachment.FileName,
		&attachment.FileExt,
		&attachment.MimeType,
		&attachment.FileSize,
		&attachment.Storage,
		&attachment.ObjectKey,
		&attachment.URL,
		&attachment.Hash,
		&attachment.Status,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
		&attachment.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if userID != nil {
		attachment.UserID = *userID
	}
	return &attachment, nil
}

func (r *PostgresAttachmentRepository) GetByHash(ctx context.Context, tenantID uint, hash string) (_ *Attachment, err error) {
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	query := `
		SELECT id, tenant_id, user_id, file_name, file_ext, mime_type, file_size, storage, object_key, url, hash, status, created_at, updated_at, is_deleted
		FROM attachments
		WHERE tenant_id = $1 AND hash = $2 AND status = 1 AND is_deleted = false
		LIMIT 1
	`
	var attachment Attachment
	var userID *uint
	err = q.QueryRow(ctx, query, tenantID, hash).Scan(
		&attachment.ID,
		&attachment.TenantID,
		&userID,
		&attachment.FileName,
		&attachment.FileExt,
		&attachment.MimeType,
		&attachment.FileSize,
		&attachment.Storage,
		&attachment.ObjectKey,
		&attachment.URL,
		&attachment.Hash,
		&attachment.Status,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
		&attachment.IsDeleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if userID != nil {
		attachment.UserID = *userID
	}
	return &attachment, nil
}

func (r *PostgresAttachmentRepository) Create(ctx context.Context, attachment *Attachment) (_ *Attachment, err error) {
	ctx, q, tx, err := db.GetTenantQuerier(ctx, r.db, attachment.TenantID)
	if err != nil {
		return nil, err
	}
	defer func() { err = db.FinishTx(ctx, tx, err) }()

	query := `
		INSERT INTO attachments (tenant_id, user_id, file_name, file_ext, mime_type, file_size, storage, object_key, url, hash, status, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, false)
		RETURNING id
	`

	now := time.Now()
	if attachment.CreatedAt.IsZero() {
		attachment.CreatedAt = now
	}
	attachment.UpdatedAt = now

	if attachment.Status == 0 {
		attachment.Status = 1
	}

	var userID *uint
	if attachment.UserID > 0 {
		userID = &attachment.UserID
	}

	err = q.QueryRow(ctx, query,
		attachment.TenantID,
		userID,
		attachment.FileName,
		attachment.FileExt,
		attachment.MimeType,
		attachment.FileSize,
		attachment.Storage,
		attachment.ObjectKey,
		attachment.URL,
		attachment.Hash,
		attachment.Status,
		attachment.CreatedAt,
		attachment.UpdatedAt,
	).Scan(&attachment.ID)

	if err != nil {
		return nil, err
	}
	return attachment, nil
}

func (r *PostgresAttachmentRepository) UpdateStatus(ctx context.Context, id uint, status int8) error {
	query := `UPDATE attachments SET status = $1, updated_at = NOW() WHERE id = $2 AND is_deleted = false`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *PostgresAttachmentRepository) Delete(ctx context.Context, id uint) error {
	query := `UPDATE attachments SET is_deleted = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

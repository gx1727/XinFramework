package asset

import (
	"context"
	"time"
)

// Attachment represents a file uploaded by a user
type Attachment struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	UserID    uint      `json:"user_id"`
	FileName  string    `json:"file_name"`
	FileExt   string    `json:"file_ext"`
	MimeType  string    `json:"mime_type"`
	FileSize  int64     `json:"file_size"`
	Storage   string    `json:"storage"`
	ObjectKey string    `json:"object_key"`
	URL       string    `json:"url"`
	Hash      string    `json:"hash"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// AttachmentRepository defines data access operations for attachments
type AttachmentRepository interface {
	GetByID(ctx context.Context, id uint) (*Attachment, error)
	GetByHash(ctx context.Context, tenantID uint, hash string) (*Attachment, error)
	Create(ctx context.Context, attachment *Attachment) (*Attachment, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	Delete(ctx context.Context, id uint) error
}

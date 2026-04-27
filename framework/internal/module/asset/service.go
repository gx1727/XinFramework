package asset

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/storage"
)

type FileService struct {
	storage storage.Storage
	repo    model.AttachmentRepository
}

func NewFileService(storage storage.Storage, repo model.AttachmentRepository) *FileService {
	return &FileService{
		storage: storage,
		repo:    repo,
	}
}

// UploadResponse returns the result of the upload
type UploadResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

// Upload handles the file upload process with deduplication and tenant isolation
func (s *FileService) Upload(ctx context.Context, tenantID uint, userID uint, file *multipart.FileHeader) (*UploadResponse, error) {
	// 1. Open file to read content
	src, err := file.Open()
	if err != nil {
		logger.Errorf("failed to open uploaded file: %v", err)
		return nil, ErrUploadFailed
	}
	defer src.Close()

	// Read content into memory to compute hash (for reasonable file sizes)
	// For very large files, streaming hash calculation should be considered
	var buf bytes.Buffer
	tee := io.TeeReader(src, &buf)

	// Compute hash
	hashWriter := sha256.New()
	if _, err := io.Copy(hashWriter, tee); err != nil {
		logger.Errorf("failed to calculate file hash: %v", err)
		return nil, ErrUploadFailed
	}
	fileHash := hex.EncodeToString(hashWriter.Sum(nil))

	// 2. 查重 (Deduplication)
	existing, err := s.repo.GetByHash(ctx, tenantID, fileHash)
	if err != nil {
		logger.Errorf("failed to query attachment by hash: %v", err)
		return nil, ErrUploadFailed
	}

	if existing != nil {
		// File already exists, reuse the existing record's storage object
		attachment := &model.Attachment{
			TenantID:  tenantID,
			UserID:    userID,
			FileName:  file.Filename,
			FileExt:   filepath.Ext(file.Filename),
			MimeType:  file.Header.Get("Content-Type"),
			FileSize:  file.Size,
			Storage:   existing.Storage,
			ObjectKey: existing.ObjectKey,
			URL:       existing.URL,
			Hash:      fileHash,
			Status:    1,
		}

		created, err := s.repo.Create(ctx, attachment)
		if err != nil {
			logger.Errorf("failed to create attachment record (dedup): %v", err)
			return nil, ErrUploadFailed
		}
		return &UploadResponse{ID: created.ID, URL: created.URL}, nil
	}

	// 3. 生成 key (Generate isolated key: tenant_id/yyyy/mm/dd/uuid.ext)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	now := time.Now()
	uuidObj, _ := uuid.NewRandom()
	objectKey := fmt.Sprintf("%d/%04d/%02d/%02d/%s%s",
		tenantID, now.Year(), now.Month(), now.Day(), uuidObj.String(), ext)

	// 4. 上传 storage
	fileReader := bytes.NewReader(buf.Bytes())
	url, err := s.storage.Put(ctx, objectKey, fileReader)
	if err != nil {
		logger.Errorf("failed to put file to storage: %v", err)
		return nil, ErrUploadFailed
	}

	// 5. 写 DB
	attachment := &model.Attachment{
		TenantID:  tenantID,
		UserID:    userID,
		FileName:  file.Filename,
		FileExt:   ext,
		MimeType:  file.Header.Get("Content-Type"),
		FileSize:  file.Size,
		Storage:   "local", // or from config
		ObjectKey: objectKey,
		URL:       url,
		Hash:      fileHash,
		Status:    1,
	}

	created, err := s.repo.Create(ctx, attachment)
	if err != nil {
		logger.Errorf("failed to create attachment record: %v", err)
		return nil, ErrUploadFailed
	}

	return &UploadResponse{ID: created.ID, URL: created.URL}, nil
}

// Delete logically deletes the attachment (soft delete)
func (s *FileService) Delete(ctx context.Context, tenantID uint, id uint) error {
	attachment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUploadFailed
	}
	if attachment == nil || attachment.TenantID != tenantID {
		return ErrFileNotFound
	}

	return s.repo.UpdateStatus(ctx, id, 0)
}

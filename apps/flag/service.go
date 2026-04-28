package flag

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/storage"
)

type Service struct {
	storage       storage.Storage
	frameRepo     *FrameRepository
	avatarRepo    *AvatarRepository
	frameCatRepo  *FrameCategoryRepository
	avatarCatRepo *AvatarCategoryRepository
}

func NewService(storage storage.Storage, frameRepo *FrameRepository, avatarRepo *AvatarRepository, frameCatRepo *FrameCategoryRepository, avatarCatRepo *AvatarCategoryRepository) *Service {
	return &Service{storage: storage, frameRepo: frameRepo, avatarRepo: avatarRepo, frameCatRepo: frameCatRepo, avatarCatRepo: avatarCatRepo}
}

type FrameCategory struct {
	ID       uint   `json:"id"`
	TenantID uint   `json:"tenant_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

type Frame struct {
	ID             uint      `json:"id"`
	TenantID       uint      `json:"tenant_id"`
	CategoryID     uint      `json:"category_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	PreviewURL     string    `json:"preview_url"`
	TemplateURL    string    `json:"template_url"`
	TemplateConfig string    `json:"template_config,omitempty"`
	Type           string    `json:"type"`
	Sort           int       `json:"sort"`
	Status         int8      `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type FrameTemplateConfig struct {
	AvatarX      int `json:"avatar_x"`
	AvatarY      int `json:"avatar_y"`
	AvatarWidth  int `json:"avatar_width"`
	AvatarHeight int `json:"avatar_height"`
	TextX        int `json:"text_x"`
	TextY        int `json:"text_y"`
	TextWidth    int `json:"text_width"`
	FontSize     int `json:"font_size"`
}

type Space struct {
	ID          uint         `json:"id"`
	TenantID    uint         `json:"tenant_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	FrameID     uint         `json:"frame_id"`
	Frame       *Frame       `json:"frame,omitempty"`
	SpaceConfig *SpaceConfig `json:"space_config,omitempty"`
	AccessType  string       `json:"access_type"`
	InviteCode  string       `json:"invite_code"`
	MaxUsage    int          `json:"max_usage"`
	UsageCount  int          `json:"usage_count"`
	Status      int8         `json:"status"`
	StartAt     *time.Time   `json:"start_at,omitempty"`
	EndAt       *time.Time   `json:"end_at,omitempty"`
}

type SpaceConfig struct {
	Fields []FieldConfig `json:"fields"`
}

type FieldConfig struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Required  bool   `json:"required"`
	Show      bool   `json:"show"`
	MaxLength int    `json:"max_length"`
}

type UserGenerated struct {
	ID          uint              `json:"id"`
	TenantID    uint              `json:"tenant_id"`
	UserID      uint              `json:"user_id"`
	SpaceID     uint              `json:"space_id"`
	FrameID     uint              `json:"frame_id"`
	SourceImage string            `json:"source_image"`
	ResultURL   string            `json:"result_url"`
	ResultKey   string            `json:"result_key"`
	FieldValues map[string]string `json:"field_values"`
	ShareText   string            `json:"share_text"`
	CreatedAt   time.Time         `json:"created_at"`
}

type GenerateResult struct {
	ID        uint   `json:"id"`
	ResultURL string `json:"result_url"`
	ShareText string `json:"share_text"`
}

type AvatarCategory struct {
	ID       uint   `json:"id"`
	TenantID uint   `json:"tenant_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Type     string `json:"type"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

type Avatar struct {
	ID           uint      `json:"id"`
	TenantID     uint      `json:"tenant_id"`
	UserID       uint      `json:"user_id"`
	CategoryID   uint      `json:"category_id"`
	Name         string    `json:"name"`
	SourceURL    string    `json:"source_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	FileSize     int64     `json:"file_size"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Type         string    `json:"type"`
	IsPublic     bool      `json:"is_public"`
	LikeCount    int       `json:"like_count"`
	ViewCount    int       `json:"view_count"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ==================== Frame CRUD ====================

func (s *Service) ListFrames(ctx context.Context, req listFramesRequest) ([]Frame, int64, error) {
	return s.frameRepo.List(ctx, req.CategoryID, req.Page, req.Size)
}

func (s *Service) GetFrame(ctx context.Context, id uint) (*Frame, error) {
	return s.frameRepo.GetByID(ctx, id)
}

func (s *Service) CreateFrame(ctx context.Context, tenantID uint, req createFrameRequest) (*Frame, error) {
	frame := &Frame{
		TenantID:    tenantID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		PreviewURL:  req.PreviewURL,
		TemplateURL: req.TemplateURL,
		Type:        req.Type,
		Sort:        req.Sort,
		Status:      1,
	}
	return s.frameRepo.Create(ctx, frame)
}

func (s *Service) UpdateFrame(ctx context.Context, tenantID uint, req updateFrameRequest) error {
	frame := &Frame{
		ID:          req.ID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		PreviewURL:  req.PreviewURL,
		TemplateURL: req.TemplateURL,
		Type:        req.Type,
		Sort:        req.Sort,
		Status:      req.Status,
	}
	return s.frameRepo.Update(ctx, frame)
}

func (s *Service) DeleteFrame(ctx context.Context, id uint) error {
	return s.frameRepo.Delete(ctx, id)
}

// ==================== Avatar CRUD ====================

func (s *Service) ListAvatars(ctx context.Context, req listAvatarsRequest) ([]Avatar, int64, error) {
	return s.avatarRepo.List(ctx, req.CategoryID, req.UserID, req.Type, req.Page, req.Size)
}

func (s *Service) GetAvatar(ctx context.Context, id uint) (*Avatar, error) {
	return s.avatarRepo.GetByID(ctx, id)
}

func (s *Service) CreateAvatar(ctx context.Context, tenantID, userID uint, req createAvatarRequest) (*Avatar, error) {
	avatar := &Avatar{
		TenantID:     tenantID,
		UserID:       userID,
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		SourceURL:    req.SourceURL,
		ThumbnailURL: req.ThumbnailURL,
		FileSize:     req.FileSize,
		Width:        req.Width,
		Height:       req.Height,
		Type:         "custom",
		IsPublic:     req.IsPublic,
		Status:       1,
	}
	return s.avatarRepo.Create(ctx, avatar)
}

func (s *Service) UpdateAvatar(ctx context.Context, tenantID uint, req updateAvatarRequest) error {
	avatar := &Avatar{
		ID:         req.ID,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		IsPublic:   req.IsPublic,
		Status:     req.Status,
	}
	return s.avatarRepo.Update(ctx, avatar)
}

func (s *Service) DeleteAvatar(ctx context.Context, tenantID, avatarID uint) error {
	return s.avatarRepo.Delete(ctx, avatarID)
}

// ==================== Frame Categories ====================

func (s *Service) ListCategories(ctx context.Context) ([]FrameCategory, error) {
	return s.frameCatRepo.List(ctx)
}

func (s *Service) GetFrameCategory(ctx context.Context, id uint) (*FrameCategory, error) {
	return s.frameCatRepo.GetByID(ctx, id)
}

func (s *Service) CreateFrameCategory(ctx context.Context, tenantID uint, req createFrameCategoryRequest) (*FrameCategory, error) {
	c := &FrameCategory{
		TenantID: tenantID,
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		Sort:     req.Sort,
		Status:   1,
	}
	return s.frameCatRepo.Create(ctx, c)
}

func (s *Service) UpdateFrameCategory(ctx context.Context, tenantID uint, req updateFrameCategoryRequest) error {
	c := &FrameCategory{
		ID:     req.ID,
		Code:   req.Code,
		Name:   req.Name,
		Type:   req.Type,
		Sort:   req.Sort,
		Status: req.Status,
	}
	return s.frameCatRepo.Update(ctx, c)
}

func (s *Service) DeleteFrameCategory(ctx context.Context, id uint) error {
	return s.frameCatRepo.Delete(ctx, id)
}

// ==================== Spaces (mock) ====================

func (s *Service) GetSpaceByCode(ctx context.Context, code string) (*Space, error) {
	if code == "test" {
		config := &SpaceConfig{
			Fields: []FieldConfig{
				{Key: "grade", Label: "届数", Required: true, Show: true, MaxLength: 20},
				{Key: "college", Label: "学院", Required: false, Show: true, MaxLength: 50},
			},
		}
		return &Space{
			ID:          1,
			TenantID:    1,
			Name:        "测试活动",
			Description: "这是一个测试活动",
			FrameID:     1,
			SpaceConfig: config,
			AccessType:  "public",
			InviteCode:  "test",
			Status:      1,
		}, nil
	}
	return nil, ErrSpaceNotFound
}

func (s *Service) CreateSpace(ctx context.Context, tenantID uint, req createSpaceRequest) (*Space, error) {
	inviteCode := generateInviteCode()

	space := &Space{
		ID:          1,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		FrameID:     req.FrameID,
		AccessType:  req.AccessType,
		InviteCode:  inviteCode,
		Status:      1,
	}

	logger.Infof("created space: %s for tenant: %d", space.Name, tenantID)
	return space, nil
}

func (s *Service) UpdateSpace(ctx context.Context, tenantID, spaceID uint, req updateSpaceRequest) error {
	logger.Infof("updated space: %d for tenant: %d", spaceID, tenantID)
	return nil
}

func (s *Service) DeleteSpace(ctx context.Context, tenantID, spaceID uint) error {
	logger.Infof("deleted space: %d for tenant: %d", spaceID, tenantID)
	return nil
}

func (s *Service) ListSpaces(ctx context.Context, tenantID uint) ([]Space, error) {
	return []Space{}, nil
}

func (s *Service) GenerateAvatar(ctx context.Context, tenantID, userID uint, req generateRequest) (*GenerateResult, error) {
	resultKey := fmt.Sprintf("flag/%d/%s.png", tenantID, uuid.New().String())

	resultURL := fmt.Sprintf("https://img.gx1727.com/%s", resultKey)

	result := &GenerateResult{
		ID:        1,
		ResultURL: resultURL,
		ShareText: fmt.Sprintf("我正在参加活动，快来一起玩！"),
	}

	logger.Infof("generated avatar for user: %d, result: %s", userID, resultKey)
	return result, nil
}

func (s *Service) ListMyAvatars(ctx context.Context, tenantID, userID uint) ([]UserGenerated, error) {
	return []UserGenerated{}, nil
}

func (s *Service) ListAvatarCategories(ctx context.Context, req listAvatarCategoriesRequest) ([]AvatarCategory, error) {
	return s.avatarCatRepo.List(ctx, req.Type)
}

func (s *Service) CreateAvatarCategory(ctx context.Context, tenantID uint, req createAvatarCategoryRequest) (*AvatarCategory, error) {
	c := &AvatarCategory{
		TenantID: tenantID,
		Code:     req.Code,
		Name:     req.Name,
		Icon:     req.Icon,
		Type:     req.Type,
		Sort:     req.Sort,
		Status:   1,
	}
	return s.avatarCatRepo.Create(ctx, c)
}

func (s *Service) UpdateAvatarCategory(ctx context.Context, tenantID uint, req updateAvatarCategoryRequest) error {
	c := &AvatarCategory{
		ID:     req.ID,
		Code:   req.Code,
		Name:   req.Name,
		Icon:   req.Icon,
		Type:   req.Type,
		Sort:   req.Sort,
		Status: req.Status,
	}
	return s.avatarCatRepo.Update(ctx, c)
}

func (s *Service) DeleteAvatarCategory(ctx context.Context, tenantID, categoryID uint) error {
	return s.avatarCatRepo.Delete(ctx, categoryID)
}

func generateInviteCode() string {
	uuidStr := uuid.New().String()
	return uuidStr[:8]
}

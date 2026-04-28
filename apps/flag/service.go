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
	storage storage.Storage
}

func NewService(storage storage.Storage) *Service {
	return &Service{storage: storage}
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
	ID             uint                 `json:"id"`
	TenantID       uint                 `json:"tenant_id"`
	CategoryID     uint                 `json:"category_id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	PreviewURL     string               `json:"preview_url"`
	TemplateURL    string               `json:"template_url"`
	TemplateConfig *FrameTemplateConfig `json:"template_config,omitempty"`
	Type           string               `json:"type"`
	Sort           int                  `json:"sort"`
	Status         int8                 `json:"status"`
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

// ==================== Avatar Category ====================

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

// ==================== Avatar ====================

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
}

func (s *Service) ListFrames(ctx context.Context, req listFramesRequest) ([]Frame, error) {
	frames := []Frame{
		{
			ID:          1,
			TenantID:    0,
			CategoryID:  1,
			Name:        "程序员专属",
			Description: "适合程序员的简约头像框",
			PreviewURL:  "/frames/preview/programmer.png",
			TemplateURL: "/frames/template/programmer.png",
			Type:        "public",
			Sort:        1,
			Status:      1,
		},
		{
			ID:          2,
			TenantID:    0,
			CategoryID:  1,
			Name:        "打工人",
			Description: "打工人的精神状态",
			PreviewURL:  "/frames/preview/worker.png",
			TemplateURL: "/frames/template/worker.png",
			Type:        "public",
			Sort:        2,
			Status:      1,
		},
		{
			ID:          3,
			TenantID:    0,
			CategoryID:  2,
			Name:        "校庆100周年",
			Description: "校庆活动专属头像框",
			PreviewURL:  "/frames/preview/school100.png",
			TemplateURL: "/frames/template/school100.png",
			Type:        "public",
			Sort:        1,
			Status:      1,
		},
	}

	if req.CategoryID > 0 {
		var filtered []Frame
		for _, f := range frames {
			if f.CategoryID == req.CategoryID {
				filtered = append(filtered, f)
			}
		}
		return filtered, nil
	}

	return frames, nil
}

func (s *Service) GetFrame(ctx context.Context, id uint) (*Frame, error) {
	frames, _ := s.ListFrames(ctx, listFramesRequest{})
	for _, f := range frames {
		if f.ID == id {
			return &f, nil
		}
	}
	return nil, ErrFrameNotFound
}

func (s *Service) ListCategories(ctx context.Context) ([]FrameCategory, error) {
	return []FrameCategory{
		{ID: 1, TenantID: 0, Code: "emotion", Name: "情绪类", Type: "emotion", Sort: 1, Status: 1},
		{ID: 2, TenantID: 0, Code: "school", Name: "学校活动", Type: "custom", Sort: 2, Status: 1},
		{ID: 3, TenantID: 0, Code: "company", Name: "企业活动", Type: "custom", Sort: 3, Status: 1},
		{ID: 4, TenantID: 0, Code: "hot", Name: "热点节日", Type: "hot", Sort: 4, Status: 1},
	}, nil
}

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
	return []AvatarCategory{
		{ID: 1, TenantID: 0, Code: "selfie", Name: "自拍头像", Icon: "/icons/selfie.png", Type: "public", Sort: 1, Status: 1},
		{ID: 2, TenantID: 0, Code: "artistic", Name: "艺术风格", Icon: "/icons/artistic.png", Type: "public", Sort: 2, Status: 1},
		{ID: 3, TenantID: 0, Code: "vintage", Name: "复古风格", Icon: "/icons/vintage.png", Type: "public", Sort: 3, Status: 1},
	}, nil
}

func (s *Service) CreateAvatarCategory(ctx context.Context, tenantID uint, req createAvatarCategoryRequest) (*AvatarCategory, error) {
	category := &AvatarCategory{
		ID:       1,
		TenantID: tenantID,
		Code:     req.Code,
		Name:     req.Name,
		Icon:     req.Icon,
		Type:     req.Type,
		Sort:     req.Sort,
		Status:   1,
	}
	logger.Infof("created avatar category: %s for tenant: %d", category.Name, tenantID)
	return category, nil
}

func (s *Service) UpdateAvatarCategory(ctx context.Context, tenantID uint, req updateAvatarCategoryRequest) error {
	logger.Infof("updated avatar category: %d for tenant: %d", req.ID, tenantID)
	return nil
}

func (s *Service) DeleteAvatarCategory(ctx context.Context, tenantID, categoryID uint) error {
	logger.Infof("deleted avatar category: %d for tenant: %d", categoryID, tenantID)
	return nil
}

func (s *Service) ListAvatars(ctx context.Context, req listAvatarsRequest) ([]Avatar, error) {
	avatars := []Avatar{
		{
			ID:           1,
			TenantID:     0,
			UserID:       1,
			CategoryID:   1,
			Name:         "我的头像1",
			SourceURL:    "/avatars/source/1.png",
			ThumbnailURL: "/avatars/thumb/1.png",
			FileSize:     102400,
			Width:        500,
			Height:       500,
			Type:         "custom",
			IsPublic:     true,
			LikeCount:    10,
			ViewCount:    100,
			Status:       1,
		},
	}
	return avatars, nil
}

func (s *Service) GetAvatar(ctx context.Context, id uint) (*Avatar, error) {
	return &Avatar{
		ID:           id,
		TenantID:     0,
		UserID:       1,
		CategoryID:   1,
		Name:         "我的头像",
		SourceURL:    "/avatars/source/1.png",
		ThumbnailURL: "/avatars/thumb/1.png",
		Type:         "custom",
		IsPublic:     true,
		Status:       1,
	}, nil
}

func (s *Service) CreateAvatar(ctx context.Context, tenantID, userID uint, req createAvatarRequest) (*Avatar, error) {
	avatar := &Avatar{
		ID:           1,
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
	logger.Infof("created avatar for user: %d, source: %s", userID, avatar.SourceURL)
	return avatar, nil
}

func (s *Service) UpdateAvatar(ctx context.Context, tenantID uint, req updateAvatarRequest) error {
	logger.Infof("updated avatar: %d for tenant: %d", req.ID, tenantID)
	return nil
}

func (s *Service) DeleteAvatar(ctx context.Context, tenantID, avatarID uint) error {
	logger.Infof("deleted avatar: %d for tenant: %d", avatarID, tenantID)
	return nil
}

func generateInviteCode() string {
	uuidStr := uuid.New().String()
	return uuidStr[:8]
}

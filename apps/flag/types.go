package flag

import "time"

// ==================== Data Models ====================

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

// ==================== Frame ====================

type listFramesRequest struct {
	CategoryID uint `form:"category_id"`
	Page       int  `form:"page,default=1"`
	Size       int  `form:"size,default=20"`
}

type getFrameRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type createFrameRequest struct {
	CategoryID  uint   `json:"category_id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	PreviewURL  string `json:"preview_url"`
	TemplateURL string `json:"template_url"`
	Type        string `json:"type"`
	Sort        int    `json:"sort"`
}

type updateFrameRequest struct {
	ID          uint   `json:"id" binding:"required"`
	CategoryID  uint   `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PreviewURL  string `json:"preview_url"`
	TemplateURL string `json:"template_url"`
	Type        string `json:"type"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type deleteFrameRequest struct {
	ID uint `uri:"id" binding:"required"`
}

// ==================== Space ====================

type getSpaceByCodeRequest struct {
	Code string `uri:"code" binding:"required"`
}

// ==================== Frame Categories ====================

type createFrameCategoryRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
	Type string `json:"type"`
	Sort int    `json:"sort"`
}

type updateFrameCategoryRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status"`
}

type deleteFrameCategoryRequest struct {
	ID uint `uri:"id" binding:"required"`
}

// ==================== Space ====================

type createSpaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	FrameID     uint   `json:"frame_id"`
	AccessType  string `json:"access_type"`
	StartAt     string `json:"start_at"`
	EndAt       string `json:"end_at"`
}

type updateSpaceRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FrameID     uint   `json:"frame_id"`
	Status      int8   `json:"status"`
}

type deleteSpaceRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type generateRequest struct {
	FrameID     uint              `json:"frame_id" binding:"required"`
	SpaceID     uint              `json:"space_id"`
	SourceImage string            `json:"source_image" binding:"required"`
	FieldValues map[string]string `json:"field_values"`
}

// ==================== Avatar Categories ====================

type listAvatarCategoriesRequest struct {
	Type string `form:"type"`
}

type createAvatarCategoryRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
	Icon string `json:"icon"`
	Type string `json:"type"`
	Sort int    `json:"sort"`
}

type updateAvatarCategoryRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Type   string `json:"type"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status"`
}

type deleteAvatarCategoryRequest struct {
	ID uint `uri:"id" binding:"required"`
}

// ==================== Avatars ====================

type listAvatarsRequest struct {
	CategoryID uint   `form:"category_id"`
	UserID     uint   `form:"user_id"`
	Type       string `form:"type"`
	Page       int    `form:"page,default=1"`
	Size       int    `form:"size,default=20"`
}

type getAvatarRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type createAvatarRequest struct {
	CategoryID   uint   `json:"category_id"`
	Name         string `json:"name"`
	SourceURL    string `json:"source_url" binding:"required"`
	ThumbnailURL string `json:"thumbnail_url"`
	FileSize     int64  `json:"file_size"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	IsPublic     bool   `json:"is_public"`
}

type updateAvatarRequest struct {
	ID         uint   `json:"id" binding:"required"`
	Name       string `json:"name"`
	CategoryID uint   `json:"category_id"`
	IsPublic   bool   `json:"is_public"`
	Status     int8   `json:"status"`
}

type deleteAvatarRequest struct {
	ID uint `uri:"id" binding:"required"`
}

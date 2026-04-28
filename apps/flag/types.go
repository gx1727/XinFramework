package flag

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrFrameNotFound  = resp.NewError(15001, "头像框不存在")
	ErrSpaceNotFound  = resp.NewError(15002, "活动空间不存在")
	ErrGenerateFailed = resp.NewError(15003, "头像生成失败")
	ErrAvatarNotFound = resp.NewError(15004, "头像不存在")
)

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

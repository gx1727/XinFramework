package flag

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ==================== Frame CRUD ====================

func (h *Handler) ListFrames(c *gin.Context) {
	var req listFramesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	frames, total, err := frameRepo.List(c.Request.Context(), req.CategoryID, req.Page, req.Size)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"list":  frames,
		"total": total,
		"page":  req.Page,
		"size":  req.Size,
	})
}

func (h *Handler) GetFrame(c *gin.Context) {
	var req getFrameRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	frame, err := frameRepo.GetByID(c.Request.Context(), req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, frame)
}

func (h *Handler) CreateFrame(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createFrameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	frame := &Frame{
		TenantID:    uc.TenantID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		PreviewURL:  req.PreviewURL,
		TemplateURL: req.TemplateURL,
		Type:        req.Type,
		Sort:        req.Sort,
		Status:      1,
	}

	result, err := frameRepo.Create(c.Request.Context(), frame)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, result)
}

func (h *Handler) UpdateFrame(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateFrameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

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

	if err := frameRepo.Update(c.Request.Context(), frame); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteFrame(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteFrameRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := frameRepo.Delete(c.Request.Context(), req.ID); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ==================== Categories ====================

func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := frameCatRepo.List(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, categories)
}

func (h *Handler) CreateFrameCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createFrameCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	cat := &FrameCategory{
		TenantID: uc.TenantID,
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		Sort:     req.Sort,
		Status:   1,
	}

	category, err := frameCatRepo.Create(c.Request.Context(), cat)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, category)
}

func (h *Handler) UpdateFrameCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateFrameCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	cat := &FrameCategory{
		ID:     req.ID,
		Code:   req.Code,
		Name:   req.Name,
		Type:   req.Type,
		Sort:   req.Sort,
		Status: req.Status,
	}

	if err := frameCatRepo.Update(c.Request.Context(), cat); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteFrameCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteFrameCategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := frameCatRepo.Delete(c.Request.Context(), req.ID); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ==================== Spaces ====================

func (h *Handler) GetSpaceByCode(c *gin.Context) {
	var req getSpaceByCodeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	// Mock implementation
	if req.Code == "test" {
		config := &SpaceConfig{
			Fields: []FieldConfig{
				{Key: "grade", Label: "届数", Required: true, Show: true, MaxLength: 20},
				{Key: "college", Label: "学院", Required: false, Show: true, MaxLength: 50},
			},
		}
		space := &Space{
			ID:          1,
			TenantID:    1,
			Name:        "测试活动",
			Description: "这是一个测试活动",
			FrameID:     1,
			SpaceConfig: config,
			AccessType:  "public",
			InviteCode:  "test",
			Status:      1,
		}
		resp.Success(c, space)
		return
	}

	resp.HandleError(c, ErrSpaceNotFound)
}

func (h *Handler) CreateSpace(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	inviteCode := generateInviteCode()

	space := &Space{
		ID:          1,
		TenantID:    uc.TenantID,
		Name:        req.Name,
		Description: req.Description,
		FrameID:     req.FrameID,
		AccessType:  req.AccessType,
		InviteCode:  inviteCode,
		Status:      1,
	}

	logger.Infof("created space: %s for tenant: %d", space.Name, uc.TenantID)
	resp.Success(c, space)
}

func (h *Handler) UpdateSpace(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	logger.Infof("updated space: %d for tenant: %d", req.ID, uc.TenantID)
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteSpace(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteSpaceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	logger.Infof("deleted space: %d for tenant: %d", req.ID, uc.TenantID)
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) ListSpaces(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	resp.Success(c, []Space{})
}

func (h *Handler) GenerateAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	resultKey := fmt.Sprintf("flag/%d/%s.png", uc.TenantID, uuid.New().String())
	resultURL := fmt.Sprintf("https://img.gx1727.com/%s", resultKey)

	result := &GenerateResult{
		ID:        1,
		ResultURL: resultURL,
		ShareText: fmt.Sprintf("我正在参加活动，快来一起玩！"),
	}

	logger.Infof("generated avatar for user: %d, result: %s", uc.UserID, resultKey)
	resp.Success(c, result)
}

func (h *Handler) ListMyAvatars(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.UserID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	resp.Success(c, []UserGenerated{})
}

// ==================== Avatar Categories ====================

func (h *Handler) ListAvatarCategories(c *gin.Context) {
	var req listAvatarCategoriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	categories, err := avatarCatRepo.List(c.Request.Context(), req.Type)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, categories)
}

func (h *Handler) CreateAvatarCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createAvatarCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	cat := &AvatarCategory{
		TenantID: uc.TenantID,
		Code:     req.Code,
		Name:     req.Name,
		Icon:     req.Icon,
		Type:     req.Type,
		Sort:     req.Sort,
		Status:   1,
	}

	category, err := avatarCatRepo.Create(c.Request.Context(), cat)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, category)
}

func (h *Handler) UpdateAvatarCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateAvatarCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	cat := &AvatarCategory{
		ID:     req.ID,
		Code:   req.Code,
		Name:   req.Name,
		Icon:   req.Icon,
		Type:   req.Type,
		Sort:   req.Sort,
		Status: req.Status,
	}

	if err := avatarCatRepo.Update(c.Request.Context(), cat); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteAvatarCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteAvatarCategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := avatarCatRepo.Delete(c.Request.Context(), req.ID); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ==================== Avatars ====================

func (h *Handler) ListAvatars(c *gin.Context) {
	var req listAvatarsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	avatars, total, err := avatarRepo.List(c.Request.Context(), req.CategoryID, req.UserID, req.Type, req.Page, req.Size)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"list":  avatars,
		"total": total,
		"page":  req.Page,
		"size":  req.Size,
	})
}

func (h *Handler) GetAvatar(c *gin.Context) {
	var req getAvatarRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	avatar, err := avatarRepo.GetByID(c.Request.Context(), req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, avatar)
}

func (h *Handler) CreateAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.UserID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	avatar := &Avatar{
		TenantID:     uc.TenantID,
		UserID:       uc.UserID,
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

	result, err := avatarRepo.Create(c.Request.Context(), avatar)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, result)
}

func (h *Handler) UpdateAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	avatar := &Avatar{
		ID:         req.ID,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		IsPublic:   req.IsPublic,
		Status:     req.Status,
	}

	if err := avatarRepo.Update(c.Request.Context(), avatar); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteAvatarRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := avatarRepo.Delete(c.Request.Context(), req.ID); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ==================== Helper Functions ====================

func generateInviteCode() string {
	uuidStr := uuid.New().String()
	return uuidStr[:8]
}

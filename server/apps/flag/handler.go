package flag

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ==================== Frame CRUD ====================

func (h *Handler) ListFrames(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	var req listFramesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	var frames []Frame
	var total int64
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		frames, total, err = frameRepo.List(ctx, req.CategoryID, req.Page, req.Size)
		return err
	})
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
	uc := xincontext.NewUserContext(c)
	var req getFrameRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	var frame *Frame
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		frame, err = frameRepo.GetByID(ctx, req.ID)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, frame)
}

func (h *Handler) CreateFrame(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req createFrameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	var result *Frame
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		result, err = frameRepo.Create(ctx, frame)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, result)
}

func (h *Handler) UpdateFrame(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req updateFrameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return frameRepo.Update(ctx, frame)
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteFrame(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req deleteFrameRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return frameRepo.Delete(ctx, req.ID)
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ==================== Categories ====================

func (h *Handler) ListFrameCategories(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	var categories []FrameCategory
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		categories, err = frameCatRepo.List(ctx)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, categories)
}

func (h *Handler) CreateFrameCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req createFrameCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	var category *FrameCategory
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		category, err = frameCatRepo.Create(ctx, cat)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, category)
}

func (h *Handler) UpdateFrameCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req updateFrameCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return frameCatRepo.Update(ctx, cat)
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteFrameCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req deleteFrameCategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return frameCatRepo.Delete(ctx, req.ID)
	})
	if err != nil {
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

	var req createSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	var req updateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	logger.Infof("updated space: %d for tenant: %d", req.ID, uc.TenantID)
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteSpace(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req deleteSpaceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	logger.Infof("deleted space: %d for tenant: %d", req.ID, uc.TenantID)
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) ListSpaces(c *gin.Context) {
	resp.Success(c, []Space{})
}

func (h *Handler) GenerateAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	resultKey := fmt.Sprintf("flag/%d/%s.png", uc.TenantID, uuid.New().String())

	baseURL := func() string {
		if cfgRef.Storage.Provider == "cos" {
			return cfgRef.Storage.CosBaseURL
		}
		return cfgRef.Storage.LocalBaseURL
	}()

	resultURL := fmt.Sprintf("%s/%s", baseURL, resultKey)

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
	uc := xincontext.NewUserContext(c)
	var categories []AvatarCategory
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		categories, err = avatarCatRepo.List(ctx)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, categories)
}

func (h *Handler) CreateAvatarCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req createAvatarCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	var category *AvatarCategory
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		category, err = avatarCatRepo.Create(ctx, cat)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, category)
}

func (h *Handler) UpdateAvatarCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req updateAvatarCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
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

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return avatarCatRepo.Update(ctx, cat)
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteAvatarCategory(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req deleteAvatarCategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return avatarCatRepo.Delete(ctx, req.ID)
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ==================== Avatars ====================

func (h *Handler) ListAvatars(c *gin.Context) {
	uc := xincontext.NewUserContext(c)
	var req listAvatarsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	var avatars []Avatar
	var total int64
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		avatars, total, err = avatarRepo.List(ctx, req.CategoryID, req.UserID, req.Type, req.Page, req.Size)
		return err
	})
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
	uc := xincontext.NewUserContext(c)
	var req getAvatarRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	var avatar *Avatar
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		avatar, err = avatarRepo.GetByID(ctx, req.ID)
		return err
	})
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
		resp.BadRequest(c, FormatValidationError(err))
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

	var result *Avatar
	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		var err error
		result, err = avatarRepo.Create(ctx, avatar)
		return err
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, result)
}

func (h *Handler) UpdateAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req updateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	avatar := &Avatar{
		ID:           req.ID,
		Name:         req.Name,
		CategoryID:   req.CategoryID,
		SourceURL:    req.SourceURL,
		ThumbnailURL: req.ThumbnailURL,
		IsPublic:     req.IsPublic,
		Status:       req.Status,
	}

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return avatarRepo.Update(ctx, avatar)
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	uc := xincontext.NewUserContext(c)

	var req deleteAvatarRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, FormatValidationError(err))
		return
	}

	err := db.RunInTenantTx(c.Request.Context(), dbPool, uc.TenantID, func(ctx context.Context) error {
		return avatarRepo.Delete(ctx, req.ID)
	})
	if err != nil {
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

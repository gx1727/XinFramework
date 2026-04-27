package flag

import (
	"github.com/gin-gonic/gin"
	xinContext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListFrames(c *gin.Context) {
	var req listFramesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	frames, err := h.svc.ListFrames(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, frames)
}

func (h *Handler) GetFrame(c *gin.Context) {
	var req getFrameRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	frame, err := h.svc.GetFrame(c.Request.Context(), req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, frame)
}

func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, categories)
}

func (h *Handler) GetSpaceByCode(c *gin.Context) {
	var req getSpaceByCodeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	space, err := h.svc.GetSpaceByCode(c.Request.Context(), req.Code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, space)
}

func (h *Handler) CreateSpace(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	space, err := h.svc.CreateSpace(c.Request.Context(), uc.TenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, space)
}

func (h *Handler) UpdateSpace(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.svc.UpdateSpace(c.Request.Context(), uc.TenantID, req.ID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteSpace(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteSpaceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.svc.DeleteSpace(c.Request.Context(), uc.TenantID, req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) ListSpaces(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	spaces, err := h.svc.ListSpaces(c.Request.Context(), uc.TenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, spaces)
}

func (h *Handler) GenerateAvatar(c *gin.Context) {
	uc := xinContext.NewUserContext(c)

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	result, err := h.svc.GenerateAvatar(c.Request.Context(), uc.TenantID, uc.UserID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, result)
}

func (h *Handler) ListMyAvatars(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.UserID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	avatars, err := h.svc.ListMyAvatars(c.Request.Context(), uc.TenantID, uc.UserID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, avatars)
}

// ==================== Avatar Categories ====================

func (h *Handler) ListAvatarCategories(c *gin.Context) {
	var req listAvatarCategoriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	categories, err := h.svc.ListAvatarCategories(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, categories)
}

func (h *Handler) CreateAvatarCategory(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createAvatarCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	category, err := h.svc.CreateAvatarCategory(c.Request.Context(), uc.TenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, category)
}

func (h *Handler) UpdateAvatarCategory(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateAvatarCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.svc.UpdateAvatarCategory(c.Request.Context(), uc.TenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteAvatarCategory(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteAvatarCategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.svc.DeleteAvatarCategory(c.Request.Context(), uc.TenantID, req.ID)
	if err != nil {
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

	avatars, err := h.svc.ListAvatars(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, avatars)
}

func (h *Handler) GetAvatar(c *gin.Context) {
	var req getAvatarRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	avatar, err := h.svc.GetAvatar(c.Request.Context(), req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, avatar)
}

func (h *Handler) CreateAvatar(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.UserID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req createAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	avatar, err := h.svc.CreateAvatar(c.Request.Context(), uc.TenantID, uc.UserID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, avatar)
}

func (h *Handler) UpdateAvatar(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req updateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.svc.UpdateAvatar(c.Request.Context(), uc.TenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "未登录")
		return
	}

	var req deleteAvatarRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.svc.DeleteAvatar(c.Request.Context(), uc.TenantID, req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/module/cms/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Ping(c *gin.Context) {
	resp.Success(c, gin.H{
		"domain": "cms",
		"status": "enabled",
	})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	ctx := context.New(c)
	userID := ctx.GetUserID()
	tenantID := ctx.GetTenantID()
	role := ctx.GetRole()

	if userID == 0 {
		resp.Error(c, 401, "unauthorized")
		return
	}

	user, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"user": gin.H{
			"id":        user.ID,
			"code":      user.Code,
			"real_name": user.RealName,
			"email":     user.Email,
			"phone":     user.Phone,
		},
		"tenant_id": tenantID,
		"role":      role,
	})
}

func (h *Handler) ListUsers(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "tenant_id is required")
		return
	}

	keyword := c.Query("keyword")
	page := 1
	pageSize := 20

	users, total, err := h.svc.ListUsers(c.Request.Context(), tenantID, keyword, page, pageSize)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"list":  users,
		"total": total,
	})
}

func (h *Handler) GetTenant(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "tenant_id is required")
		return
	}

	tenant, err := h.svc.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, tenant)
}

// ============ CMS Posts ============

func (h *Handler) ListPosts(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	if tenantID == 0 {
		resp.BadRequest(c, "tenant_id is required")
		return
	}

	keyword := c.Query("keyword")
	statusStr := c.Query("status")
	var status *int16
	if statusStr != "" {
		s, err := strconv.ParseInt(statusStr, 10, 16)
		if err == nil {
			ss := int16(s)
			status = &ss
		}
	}

	page := 1
	size := 20

	posts, total, err := h.svc.ListPosts(c.Request.Context(), tenantID, keyword, status, page, size)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"list":  posts,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h *Handler) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid post id")
		return
	}

	post, err := h.svc.GetPost(c.Request.Context(), uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, post)
}

func (h *Handler) CreatePost(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	if tenantID == 0 {
		resp.BadRequest(c, "tenant_id is required")
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
		Status  int16  `json:"status" binding:"required,oneof=-1 0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	post, err := h.svc.CreatePost(c.Request.Context(), tenantID, req.Title, req.Content, req.Status)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, post)
}

func (h *Handler) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid post id")
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
		Status  int16  `json:"status" binding:"required,oneof=-1 0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.svc.UpdatePost(c.Request.Context(), uint(id), req.Title, req.Content, req.Status); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid post id")
		return
	}

	if err := h.svc.DeletePost(c.Request.Context(), uint(id)); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

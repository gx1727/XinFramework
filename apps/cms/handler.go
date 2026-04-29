package cms

import (
	"strconv"

	"github.com/gin-gonic/gin"
	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/repository"
	"gx1727.com/xin/framework/pkg/resp"
)

// Handler HTTP 处理器
type Handler struct {
	// 直接依赖框架的 Provider，取代自己手写的 Repository
	provider repository.Provider
}

// NewHandler 创建 Handler 实例
func NewHandler() *Handler {
	// 启动时从全局获取已注入的核心能力 Provider
	return &Handler{provider: repository.Get()}
}

// ============ Ping ============

func (h *Handler) Ping(c *gin.Context) {
	resp.Success(c, gin.H{
		"domain": "cms",
		"status": "enabled",
	})
}

// ============ User ============

func (h *Handler) GetCurrentUser(c *gin.Context) {
	xc := xincontext.New(c)
	userID := xc.GetUserID()
	tenantID := xc.GetTenantID()
	role := xc.GetRole()

	if userID == 0 {
		resp.Error(c, 401, "unauthorized")
		return
	}

	user, err := h.provider.User().GetByID(c.Request.Context(), userID)
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
	xc := xincontext.New(c)
	tenantID := xc.GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "tenant_id is required")
		return
	}

	keyword := c.Query("keyword")
	page := 1
	pageSize := 20

	users, total, err := h.provider.User().List(c.Request.Context(), tenantID, keyword, page, pageSize)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"list":  users,
		"total": total,
	})
}

// ============ Tenant ============

func (h *Handler) GetTenant(c *gin.Context) {
	xc := xincontext.New(c)
	tenantID := xc.GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "tenant_id is required")
		return
	}

	tenant, err := h.provider.Tenant().GetByID(c.Request.Context(), tenantID)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, tenant)
}

// ============ CMS Posts ============

func (h *Handler) ListPosts(c *gin.Context) {
	xc := xincontext.New(c)
	tenantID := xc.GetTenantID()
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

	posts, total, err := h.provider.CmsPost().List(c.Request.Context(), tenantID, keyword, status, page, size)
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

	post, err := h.provider.CmsPost().GetByID(c.Request.Context(), uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, post)
}

func (h *Handler) CreatePost(c *gin.Context) {
	xc := xincontext.New(c)
	tenantID := xc.GetTenantID()
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

	post, err := h.provider.CmsPost().Create(c.Request.Context(), tenantID, req.Title, req.Content, req.Status)
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

	if err := h.provider.CmsPost().Update(c.Request.Context(), uint(id), req.Title, req.Content, req.Status); err != nil {
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

	if err := h.provider.CmsPost().Delete(c.Request.Context(), uint(id)); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

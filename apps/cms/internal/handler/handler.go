package handler

import (
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

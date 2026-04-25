package user

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user": gin.H{
			"id":        result.User.ID,
			"tenant_id": result.User.TenantID,
			"code":      result.User.Code,
			"role":      result.User.Role,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.svc.Logout(context.New(c).GetSessionID()); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user": gin.H{
			"id":        result.User.ID,
			"tenant_id": result.User.TenantID,
			"code":      result.User.Code,
			"role":      result.User.Role,
		},
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Refresh(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
	})
}

package auth

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler() *Handler {
	return &Handler{svc: NewService()}
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Login(req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"token": result.Token,
		"user": gin.H{
			"id":        result.User.ID,
			"tenant_id": result.User.TenantID,
			"code":      result.User.Code,
			"role":      result.User.Role,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	sessionID, _ := c.Get("session_id")
	sid, _ := sessionID.(string)
	if err := h.svc.Logout(sid); err != nil {
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
	result, err := h.svc.Register(req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"token": result.Token,
		"user": gin.H{
			"id":        result.User.ID,
			"tenant_id": result.User.TenantID,
			"code":      result.User.Code,
			"role":      result.User.Role,
		},
	})
}

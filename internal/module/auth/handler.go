package auth

import (
	"errors"

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
		resp.BadRequest(c, "invalid request body")
		return
	}
	result, err := h.svc.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAccountOrPassword):
			resp.Unauthorized(c, "invalid account or password")
		case errors.Is(err, ErrTenantBindingNotFound):
			resp.Forbidden(c, "user is not bound to any tenant")
		case errors.Is(err, ErrUserDisabled):
			resp.Forbidden(c, "user is disabled")
		case errors.Is(err, ErrSessionCreateFailed):
			resp.ServerError(c, "create session failed")
		case errors.Is(err, ErrBackendUnavailable):
			resp.ServerError(c, "backend not initialized")
		default:
			resp.ServerError(c, "generate token failed")
		}
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
	auth := c.GetHeader("Authorization")
	if auth == "" {
		resp.Unauthorized(c, "unauthorized")
		return
	}
	if err := h.svc.Logout(auth); err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			resp.Unauthorized(c, "invalid token")
		case errors.Is(err, ErrSessionRevokeFailed):
			resp.ServerError(c, "revoke session failed")
		case errors.Is(err, ErrBackendUnavailable):
			resp.ServerError(c, "backend not initialized")
		default:
			resp.ServerError(c, "logout failed")
		}
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

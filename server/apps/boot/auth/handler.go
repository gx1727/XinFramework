package auth

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

// serializeAuthUser 把 auth.User 拍平成 gin.H，供 login/register 响应复用。
func serializeAuthUser(u User) gin.H {
	return gin.H{
		"id":            u.ID,
		"tenant_id":     u.TenantID,
		"code":          u.Code,
		"role":          u.Role,
		"nickname":      u.Nickname,
		"real_name":     u.RealName,
		"avatar":        u.Avatar,
		"email":         u.Email,
		"platform_roles": u.PlatformRoles,
	}
}

// TenantLogin 租户域登录（业务用户登录入口）。
// POST /auth/tenant-login
// 请求：{ account, password, tenant_id }
func (h *Handler) TenantLogin(c *gin.Context) {
	var req tenantLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	if req.TenantID == 0 {
		resp.HandleError(c, ErrTenantRequired)
		return
	}
	result, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"scope":        result.Scope,
		"token":        result.Token,
		"refresh_token": result.RefreshToken,
		"user":         serializeAuthUser(result.User),
	})
}

// PlatformLogin 平台域登录（super_admin 登录入口）。
// POST /auth/platform-login
// 请求：{ account, password }    ← 无 tenant_id
func (h *Handler) PlatformLogin(c *gin.Context) {
	var req platformLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.PlatformLogin(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"scope":        result.Scope,
		"token":        result.Token,
		"refresh_token": result.RefreshToken,
		"user":         serializeAuthUser(result.User),
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
		"scope":        result.Scope,
		"token":        result.Token,
		"refresh_token": result.RefreshToken,
		"user":         serializeAuthUser(result.User),
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
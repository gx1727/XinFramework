package auth

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/framework/pkg/xincontext"
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
		"id":             u.ID,
		"tenant_id":      u.TenantID,
		"code":           u.Code,
		"role":           u.Role,
		"nickname":       u.Nickname,
		"real_name":      u.RealName,
		"avatar":         u.Avatar,
		"email":          u.Email,
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
		"scope":         result.Scope,
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          serializeAuthUser(result.User),
	})
}

// LoginPrecheck 登录前置检查入口（路径 B 多身份账号支持）。
// POST /auth/login-precheck
// 请求：{ account, password }
// 响应：{ account_id, platform_available, platform_roles, tenant_identities: [...] }
//
// 前端流程：
//  1. 用户提交账号密码
//  2. 调本接口拿到所有可选身份
//  3. UI 列出 tenant_identities 让用户选 + 显示 platform_available 入口
//  4. 用户选择后调 /auth/select-tenant 或 /auth/platform-login 签发 token
//
// 单身份账号可以直接调 /auth/tenant-login 跳过本接口。
func (h *Handler) LoginPrecheck(c *gin.Context) {
	var req loginPrecheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.LoginPrecheck(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, result)
}

// SelectTenant 选择 tenant 身份签发 token。
// POST /auth/select-tenant
// 请求：{ account, password, tenant_id }
// 响应：完整登录响应（与 /auth/tenant-login 等价）
//
// 语义：与 /auth/tenant-login 完全等价，仅作为"precheck → 选择 → 登录"流程
// 的语义化入口。复用 Service.Login 实现。
func (h *Handler) SelectTenant(c *gin.Context) {
	var req tenantLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	if req.TenantID == 0 {
		resp.HandleError(c, ErrTenantRequired)
		return
	}
	result, err := h.svc.SelectTenant(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"scope":         result.Scope,
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          serializeAuthUser(result.User),
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
		"scope":         result.Scope,
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          serializeAuthUser(result.User),
	})
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.svc.Logout(string(xincontext.New(c).GetSessionID())); err != nil {
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
		"scope":         result.Scope,
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          serializeAuthUser(result.User),
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

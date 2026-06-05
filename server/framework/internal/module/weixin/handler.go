package weixin

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Login 小程序登录
// POST /api/v1/weixin/login
func (h *Handler) Login(c *gin.Context) {
	var req Code2SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	result, err := h.svc.LoginByWeChat(c.Request.Context(), req.Code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"is_new_user":   result.IsNewUser,
		"user": gin.H{
			"id":        result.User.ID,
			"openid":    result.User.OpenID,
			"unionid":   result.User.UnionID,
			"phone":     result.User.Phone,
			"tenant_id": result.User.TenantID,
			"code":      result.User.Code,
			"role":      result.User.Role,
			"status":    result.User.Status,
		},
	})
}

// GetPhoneNumber 获取用户手机号
// POST /api/v1/weixin/phone
func (h *Handler) GetPhoneNumber(c *gin.Context) {
	var req PhoneNumberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	result, err := h.svc.GetPhoneNumber(c.Request.Context(), req.Code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{
		"phone_number":      result.PhoneInfo.PhoneNumber,
		"pure_phone_number": result.PhoneInfo.PurePhoneNumber,
		"country_code":      result.PhoneInfo.CountryCode,
	})
}

// BindPhone 绑定用户手机号
// POST /api/v1/weixin/bind-phone
func (h *Handler) BindPhone(c *gin.Context) {
	var req PhoneNumberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	userID := c.GetUint("user_id")
	if userID == 0 {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	phone, err := h.svc.UpdatePhoneByWeChat(c.Request.Context(), userID, req.Code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"phone": phone})
}

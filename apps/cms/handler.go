package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/repository"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(c *gin.Context) {
	resp.Success(c, gin.H{
		"domain": "cms",
		"status": "enabled",
		"config": moduleCfg,
	})
}

// GetUserByID 使用 repository 查询用户
func (h *Handler) GetUserByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		resp.Error(c, 401, "unauthorized")
		return
	}

	ctx := c.Request.Context()

	// 使用 framework 提供的 repository
	user, err := repository.User().GetByID(ctx, userID)
	if err != nil {
		if err == model.ErrUserNotFound {
			resp.Error(c, 404, "user not found")
			return
		}
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, user)
}

// ListTenantUsers 使用 repository 列出用户
func (h *Handler) ListTenantUsers(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		resp.Error(c, 400, "tenant_id is required")
		return
	}

	ctx := c.Request.Context()

	users, total, err := repository.User().List(ctx, tenantID, "", 1, 20)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"list":  users,
		"total": total,
	})
}

// GetTenant 使用 repository 获取租户
func (h *Handler) GetTenant(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		resp.Error(c, 400, "tenant_id is required")
		return
	}

	ctx := c.Request.Context()

	tenant, err := repository.Tenant().GetByID(ctx, tenantID)
	if err != nil {
		if err == model.ErrTenantNotFound {
			resp.Error(c, 404, "tenant not found")
			return
		}
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, tenant)
}

// SearchUsers 使用 repository 按关键字搜索
func (h *Handler) SearchUsers(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	keyword := c.Query("keyword")

	ctx := c.Request.Context()

	users, total, err := repository.User().List(ctx, tenantID, keyword, 1, 10)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"list":    users,
		"total":   total,
		"keyword": keyword,
	})
}

// GetCurrentUser 获取当前登录用户信息
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	tenantID := c.GetUint("tenant_id")
	role := c.GetString("role")

	if userID == 0 {
		resp.Error(c, 401, "unauthorized")
		return
	}

	ctx := c.Request.Context()

	// 使用 repository 获取完整用户信息
	user, err := repository.User().GetByID(ctx, userID)
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

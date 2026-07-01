package cms

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/extapi"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	pkgrbac "gx1727.com/xin/framework/pkg/tenant/auth"
	"gx1727.com/xin/framework/pkg/xincontext"
)

// Handler HTTP 处理器
type Handler struct {
	posts CmsPostRepository
	// 跨域能力通过 plugin.Reader（AppContext）直接拿 repo——不再走全局 provider。
	ctx plugin.Reader
}

// NewHandler 创建 Handler 实例
func NewHandler(pool *pgxpool.Pool, ctx plugin.Reader) *Handler {
	return &Handler{
		posts: NewCmsPostRepository(pool),
		ctx:   ctx,
	}
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

	if h.ctx == nil || h.ctx.UserRepo() == nil {
		resp.Error(c, 500, "user module not loaded — register apps/tenant/user in main.go")
		return
	}
	u, err := h.ctx.UserRepo().GetByID(c.Request.Context(), userID)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"user": gin.H{
			"id":        u.ID,
			"code":      u.Code,
			"real_name": u.RealName,
			"email":     u.Email,
			"phone":     u.Phone,
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

	if h.ctx == nil || h.ctx.UserRepo() == nil {
		resp.Error(c, 500, "user module not loaded — register apps/tenant/user in main.go")
		return
	}

	keyword := c.Query("keyword")
	page := 1
	pageSize := 20

	users, total, err := h.ctx.UserRepo().List(c.Request.Context(), tenantID, keyword, page, pageSize)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	// pkgrbac.User 与 extapi.User 字段兼容（org_id/org_name 在 DTO 上不出现），转 DTO 给前端。
	out := make([]extapi.User, len(users))
	for i := range users {
		out[i] = toExtAPIUser(&users[i])
	}
	resp.Success(c, gin.H{
		"list":  out,
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

	if h.ctx == nil || h.ctx.TenantRepo() == nil {
		resp.Error(c, 500, "tenant module not loaded — register apps/sys/tenant in main.go")
		return
	}
	t, err := h.ctx.TenantRepo().GetByID(c.Request.Context(), tenantID)
	if err != nil {
		resp.Error(c, 500, err.Error())
		return
	}

	resp.Success(c, toExtAPITenant(t))
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

	posts, total, err := h.posts.List(c.Request.Context(), tenantID, keyword, status, page, size)
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

	post, err := h.posts.GetByID(c.Request.Context(), uint(id))
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

	post, err := h.posts.Create(c.Request.Context(), tenantID, req.Title, req.Content, req.Status)
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

	if err := h.posts.Update(c.Request.Context(), uint(id), req.Title, req.Content, req.Status); err != nil {
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

	if err := h.posts.Delete(c.Request.Context(), uint(id)); err != nil {
		resp.HandleError(c, mapRepoError(err))
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

// ============ DTO 转换 ============
//
// pkgrbac.User / pkgtenant.TenantRecord 是 framework 内部窄接口（无 json tag 或 tag 不全），
// 给前端返回 JSON 时转成 extapi.User / extapi.Tenant DTO。字段集兼容，1:1 copy。

func toExtAPIUser(u *pkgrbac.User) extapi.User {
	return extapi.User{
		ID:        u.ID,
		TenantID:  u.TenantID,
		AccountID: u.AccountID,
		Code:      u.Code,
		Nickname:  u.Nickname,
		Status:    u.Status,
		RealName:  u.RealName,
		Avatar:    u.Avatar,
		Phone:     u.Phone,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toExtAPITenant(t *pkgtenant.TenantRecord) extapi.Tenant {
	return extapi.Tenant{
		ID:        t.ID,
		Code:      t.Code,
		Name:      t.Name,
		Status:    t.Status,
		Contact:   t.Contact,
		Phone:     t.Phone,
		Email:     t.Email,
		Province:  t.Province,
		City:      t.City,
		Area:      t.Area,
		Address:   t.Address,
		Config:    t.Config,
		Dashboard: t.Dashboard,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

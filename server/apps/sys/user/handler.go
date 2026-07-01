package sysuser

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// operatorID 从 xc（XinContext）里取——Auth 中间件已注入。
// 若取不到则记 0（系统操作）。sys 域所有操作都假定 super_admin 身份。
func operatorID(c *gin.Context) uint {
	if v, ok := c.Get("xc"); ok {
		if xc, ok := v.(interface{ GetUserID() uint }); ok {
			return xc.GetUserID()
		}
	}
	// 兜底：直接从 context 取 user_id（Auth 中间件设置）
	if uid, exists := c.Get("user_id"); exists {
		if u, ok := uid.(uint); ok {
			return u
		}
	}
	return 0
}

func (h *Handler) List(c *gin.Context) {
	var q ListQuery
	_ = c.ShouldBindQuery(&q)
	list, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Paginate(c, total, list)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSysUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	out, err := h.svc.Create(c.Request.Context(), req, operatorID(c))
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req UpdateSysUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req, operatorID(c))
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var body struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, body.Status, operatorID(c)); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, operatorID(c)); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) AssignRoles(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req AssignRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	if err := h.svc.AssignRoles(c.Request.Context(), id, req.RoleIDs); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func parseIDParam(c *gin.Context, param string) (uint, error) {
	str := c.Param(param)
	if str == "" {
		return 0, strconv.ErrSyntax
	}
	n, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

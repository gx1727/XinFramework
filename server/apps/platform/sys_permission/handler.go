package syspermission

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

func operatorID(c *gin.Context) uint {
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
	var req CreateSysPermissionReq
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
	var req UpdateSysPermissionReq
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

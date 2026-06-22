package tenants

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

func (h *Handler) Create(c *gin.Context) {
	var req CreateTenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, result)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req UpdateTenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, result)
}

// UpdateStatus 启用 / 禁用租户。状态字段独立接口，权限可细分到 ActUpdate　
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req UpdateTenantStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, result)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// Purge 硬删租户。不可逆操作，必须兀Delete 冀Purge　
// 审计会自动写 db_logs，无需手动调用　
func (h *Handler) Purge(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	result, err := h.svc.Purge(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{
		"tenant_id":     result.TenantID,
		"code":          result.Code,
		"tables_purged": len(result.Tables),
		"tables":        result.Tables,
	})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	result, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, result)
}

func (h *Handler) List(c *gin.Context) {
	var req ListTenantReq
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Paginate(c, total, list)
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

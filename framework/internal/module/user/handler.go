package user

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

func (h *Handler) List(c *gin.Context) {
	var req listRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	list, total, err := h.svc.List(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, listResponse{
		List:  list,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

func (h *Handler) Get(c *gin.Context) {
	var req getRequest
	if err := c.ShouldBindUri(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	info, err := h.svc.Get(c.Request.Context(), tenantID, req.ID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, info)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	if err := h.svc.UpdateStatus(c.Request.Context(), tenantID, req.ID, req.Status); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

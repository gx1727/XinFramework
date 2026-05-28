package resource

import (
	"strconv"

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
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "invalid parameters")
		return
	}

	list, total, err := h.svc.List(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, listResponse{List: list, Total: total})
}

func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid resource id")
		return
	}

	r, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, r)
}

func (h *Handler) Create(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	r, err := h.svc.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, r)
}

func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid resource id")
		return
	}

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body")
		return
	}

	r, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, r)
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid resource id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) GetByMenu(c *gin.Context) {
	menuIDStr := c.Param("menu_id")
	menuID, err := strconv.ParseUint(menuIDStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid menu id")
		return
	}

	list, err := h.svc.GetByMenu(c.Request.Context(), uint(menuID))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, list)
}

// GetMyResources 查询当前用户在指定菜单下可访问的资源（前端按钮权限）
func (h *Handler) GetMyResources(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	userID := ctx.GetUserID()

	menuIDStr := c.Query("menu_id")
	menuID, err := strconv.ParseUint(menuIDStr, 10, 64)
	if err != nil || menuID == 0 {
		resp.BadRequest(c, "invalid or missing menu_id")
		return
	}

	list, err := h.svc.GetUserResourcesByMenu(c.Request.Context(), tenantID, userID, uint(menuID))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, list)
}

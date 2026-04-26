package menu

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

func (h *Handler) Create(c *gin.Context) {
	var req CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	xc := context.New(c)
	tenantID := xc.GetTenantID()
	if tenantID == 0 {
		resp.BadRequest(c, "missing tenant_id")
		return
	}

	result, err := h.svc.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.Error(c, MapRespError(err), err.Error())
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

	var req UpdateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	result, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		resp.Error(c, MapRespError(err), err.Error())
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
		resp.Error(c, MapRespError(err), err.Error())
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}

	result, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.Error(c, MapRespError(err), err.Error())
		return
	}
	resp.Success(c, result)
}

func (h *Handler) List(c *gin.Context) {
	var req ListMenuReq
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	xc := context.New(c)
	tenantID := xc.GetTenantID()
	if tenantID == 0 {
		resp.BadRequest(c, "missing tenant_id")
		return
	}

	list, total, err := h.svc.List(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Paginate(c, total, list)
}

func (h *Handler) Tree(c *gin.Context) {
	xc := context.New(c)
	tenantID := xc.GetTenantID()
	if tenantID == 0 {
		resp.BadRequest(c, "missing tenant_id")
		return
	}

	tree, err := h.svc.Tree(c.Request.Context(), tenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, tree)
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

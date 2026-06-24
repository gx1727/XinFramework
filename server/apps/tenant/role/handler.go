package role

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/xincontext"

	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	ctx := xincontext.New(c)
	tenantID := ctx.GetTenantID()

	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "invalid parameters: "+err.Error())
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
		resp.BadRequest(c, "invalid role id")
		return
	}

	role, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, role)
}

func (h *Handler) Create(c *gin.Context) {
	ctx := xincontext.New(c)
	tenantID := ctx.GetTenantID()

	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	role, err := h.svc.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, role)
}

func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	role, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, role)
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// Patch 局部更新角色；body 中未提供的字段保持原值
func (h *Handler) Patch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	var req PatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	role, err := h.svc.Patch(c.Request.Context(), uint(id), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, role)
}

func (h *Handler) GetDataScopes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	ds, err := h.svc.GetDataScopes(c.Request.Context(), uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, ds)
}

func (h *Handler) UpdateDataScopes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	var req UpdateDataScopesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	if err := h.svc.UpdateDataScopes(c.Request.Context(), uint(id), req); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) GetMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	menus, err := h.svc.GetMenus(c.Request.Context(), uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, menus)
}

func (h *Handler) AssignMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	var req AssignMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	if err := h.svc.AssignMenus(c.Request.Context(), uint(id), req); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func handleError(c *gin.Context, err error) {
	switch err {
	case ErrRoleNotFoundDB:
		resp.NotFound(c, "role not found")
	case ErrRoleCodeExists:
		resp.Error(c, 400, "role code already exists")
	case ErrCannotDeleteAdmin:
		resp.Error(c, 400, "cannot delete admin role")
	default:
		resp.ServerError(c, err.Error())
	}
}

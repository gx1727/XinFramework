package permission

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

func (h *Handler) GetPermissions(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	perms, err := h.svc.GetPermissions(c.Request.Context(), uint(roleID))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, perms)
}

func (h *Handler) AssignPermissions(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	var req AssignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body")
		return
	}

	if err := h.svc.AssignPermissions(c.Request.Context(), tenantID, uint(roleID), req); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) GetMenus(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	menus, err := h.svc.GetMenus(c.Request.Context(), uint(roleID))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"list": menus})
}

func (h *Handler) GetResources(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "invalid role id")
		return
	}

	resources, err := h.svc.GetResources(c.Request.Context(), uint(roleID))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"list": resources})
}

// Package dict ?? handler
package dict

import (
	"errors"
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

// List ????????
func (h *Handler) List(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req listRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "????????")
		return
	}

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

// Get ????
func (h *Handler) Get(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "?????ID")
		return
	}

	d, err := h.svc.Get(c.Request.Context(), tenantID, id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// Create ????
func (h *Handler) Create(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "????????")
		return
	}

	d, err := h.svc.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, ErrDictCodeExists) {
			resp.Error(c, 409, "???????")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// Update ????
func (h *Handler) Update(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "?????ID")
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "????????")
		return
	}

	d, err := h.svc.Update(c.Request.Context(), tenantID, id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// Delete ??????????????????????
func (h *Handler) Delete(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "?????ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), tenantID, id); err != nil {
		if errors.Is(err, ErrDictHasItems) {
			resp.Error(c, 409, err.Error())
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ListItems ???????????
func (h *Handler) ListItems(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req listItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "????????")
		return
	}

	items, err := h.svc.ListItems(c.Request.Context(), tenantID, req.DictID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"list": items, "total": int64(len(items))})
}

// CreateItem ?????????
func (h *Handler) CreateItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "?????ID")
		return
	}

	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "????????")
		return
	}

	item, err := h.svc.CreateItem(c.Request.Context(), tenantID, dictID, req)
	if err != nil {
		if errors.Is(err, ErrDictItemCodeExists) {
			resp.Error(c, 409, "????????")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// UpdateItem ?????
func (h *Handler) UpdateItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	itemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "??????ID")
		return
	}

	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "????????")
		return
	}

	if err := h.svc.UpdateItem(c.Request.Context(), tenantID, itemID, req); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// DeleteItem ?????
func (h *Handler) DeleteItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	itemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "??????ID")
		return
	}

	if err := h.svc.DeleteItem(c.Request.Context(), tenantID, itemID); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

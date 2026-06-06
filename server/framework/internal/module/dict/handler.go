// Package dict 数据字典 HTTP 处理器
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

// List 字典列表（分页 + 关键字）
func (h *Handler) List(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req listRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
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

// Get 获取单个字典
func (h *Handler) Get(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	d, err := h.svc.Get(c.Request.Context(), tenantID, id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// Create 新建字典
func (h *Handler) Create(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	d, err := h.svc.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, ErrDictCodeExists) {
			resp.Error(c, 409, "字典编码已存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// Update 更新字典基础信息
func (h *Handler) Update(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	d, err := h.svc.Update(c.Request.Context(), tenantID, id, req)
	if err != nil {
		if errors.Is(err, ErrDictNotFound) {
			resp.Error(c, 404, "字典不存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// Delete 删除字典（若有字典项则拒绝并提示）
func (h *Handler) Delete(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), tenantID, id); err != nil {
		if errors.Is(err, ErrDictNotFound) {
			resp.Error(c, 404, "字典不存在")
			return
		}
		if errors.Is(err, ErrDictHasItems) {
			resp.Error(c, 409, err.Error())
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ListItems 列出某字典下的字典项
func (h *Handler) ListItems(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	items, err := h.svc.ListItems(c.Request.Context(), tenantID, dictID)
	if err != nil {
		if errors.Is(err, ErrDictNotFound) {
			resp.Error(c, 404, "字典不存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"list": items, "total": int64(len(items))})
}

// CreateItem 在指定字典下新增字典项
func (h *Handler) CreateItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	item, err := h.svc.CreateItem(c.Request.Context(), tenantID, dictID, req)
	if err != nil {
		if errors.Is(err, ErrDictNotFound) {
			resp.Error(c, 404, "字典不存在")
			return
		}
		if errors.Is(err, ErrDictItemCodeExists) {
			resp.Error(c, 409, "字典项编码已存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// UpdateItem 更新字典项
func (h *Handler) UpdateItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	itemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}

	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.svc.UpdateItem(c.Request.Context(), tenantID, itemID, req); err != nil {
		if errors.Is(err, ErrDictItemNotFound) {
			resp.Error(c, 404, "字典项不存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// DeleteItem 删除字典项
func (h *Handler) DeleteItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	itemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}

	if err := h.svc.DeleteItem(c.Request.Context(), tenantID, itemID); err != nil {
		if errors.Is(err, ErrDictItemNotFound) {
			resp.Error(c, 404, "字典项不存在")
			return
		}
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

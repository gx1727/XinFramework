// Package config 通用配置 - HTTP 处理器
package config

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

// =============== Group ===============

// ListGroups 列出当前租户的所有配置分组
func (h *Handler) ListGroups(c *gin.Context) {
	uc := context.NewUserContext(c)
	groups, err := h.svc.ListGroups(c.Request.Context(), uc.TenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"list": groups, "total": len(groups)})
}

// CreateGroup 新建分组
func (h *Handler) CreateGroup(c *gin.Context) {
	uc := context.NewUserContext(c)
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	g, err := h.svc.CreateGroup(c.Request.Context(), uc.TenantID, req)
	if err != nil {
		if errors.Is(err, ErrGroupCodeExists) {
			resp.Error(c, 409, "分组编码已存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, g)
}

// UpdateGroup 更新分组
func (h *Handler) UpdateGroup(c *gin.Context) {
	uc := context.NewUserContext(c)
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分组ID")
		return
	}
	var req updateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	g, err := h.svc.UpdateGroup(c.Request.Context(), uc.TenantID, id, req)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			resp.Error(c, 404, "分组不存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, g)
}

// DeleteGroup 删除分组
func (h *Handler) DeleteGroup(c *gin.Context) {
	uc := context.NewUserContext(c)
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分组ID")
		return
	}
	if err := h.svc.DeleteGroup(c.Request.Context(), uc.TenantID, id); err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			resp.Error(c, 404, "分组不存在")
			return
		}
		if errors.Is(err, ErrGroupIsSystem) {
			resp.Error(c, 403, "系统预置分组不可删除")
			return
		}
		if errors.Is(err, ErrGroupHasItems) {
			resp.Error(c, 409, err.Error())
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// =============== Item ===============

// ListItemsByGroup 列出某分组下的所有项
func (h *Handler) ListItemsByGroup(c *gin.Context) {
	uc := context.NewUserContext(c)
	groupID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分组ID")
		return
	}
	items, err := h.svc.ListItemsByGroup(c.Request.Context(), uc.TenantID, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			resp.Error(c, 404, "分组不存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"list": items, "total": len(items)})
}

// ListAllItems 列出当前租户所有项（用于批量编辑场景）
func (h *Handler) ListAllItems(c *gin.Context) {
	uc := context.NewUserContext(c)
	items, err := h.svc.ListItemsByTenant(c.Request.Context(), uc.TenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"list": items, "total": len(items)})
}

// CreateItem 在指定分组下新增项
func (h *Handler) CreateItem(c *gin.Context) {
	uc := context.NewUserContext(c)
	groupID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分组ID")
		return
	}
	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	item, err := h.svc.CreateItem(c.Request.Context(), uc.TenantID, groupID, req)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			resp.Error(c, 404, "分组不存在")
			return
		}
		if errors.Is(err, ErrItemKeyExists) {
			resp.Error(c, 409, "项 key 已存在")
			return
		}
		if errors.Is(err, ErrInvalidValueForType) || errors.Is(err, ErrValueNotInOptions) || errors.Is(err, ErrInvalidItemType) {
			resp.Error(c, 400, err.Error())
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// UpdateItem 更新项（值 + 元数据）
func (h *Handler) UpdateItem(c *gin.Context) {
	uc := context.NewUserContext(c)
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的项ID")
		return
	}
	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	item, err := h.svc.UpdateItem(c.Request.Context(), uc.TenantID, id, req)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			resp.Error(c, 404, "项不存在")
			return
		}
		if errors.Is(err, ErrItemIsReadonly) {
			resp.Error(c, 403, "该项只读，不可修改")
			return
		}
		if errors.Is(err, ErrInvalidValueForType) || errors.Is(err, ErrValueNotInOptions) {
			resp.Error(c, 400, err.Error())
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// ResetItem 恢复默认
func (h *Handler) ResetItem(c *gin.Context) {
	uc := context.NewUserContext(c)
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的项ID")
		return
	}
	item, err := h.svc.ResetItem(c.Request.Context(), uc.TenantID, id)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			resp.Error(c, 404, "项不存在")
			return
		}
		if errors.Is(err, ErrItemIsReadonly) {
			resp.Error(c, 403, "该项只读")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// DeleteItem 删除项
func (h *Handler) DeleteItem(c *gin.Context) {
	uc := context.NewUserContext(c)
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的项ID")
		return
	}
	if err := h.svc.DeleteItem(c.Request.Context(), uc.TenantID, id); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			resp.Error(c, 404, "项不存在")
			return
		}
		if errors.Is(err, ErrItemIsSystem) {
			resp.Error(c, 403, "系统预置项不可删除")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// =============== Public 公共读 ===============

// GetPublic 公共读：取当前租户某分组下的所有 is_public 项，扁平化为 key→value
func (h *Handler) GetPublic(c *gin.Context) {
	groupCode := c.Query("group")
	if groupCode == "" {
		resp.BadRequest(c, "缺少 group 参数")
		return
	}
	tenantID := resolveTenantID(c)
	values, err := h.svc.GetPublicByGroup(c.Request.Context(), tenantID, groupCode)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, publicConfigResponse{
		Group:  groupCode,
		Values: values,
	})
}

// resolveTenantID 公共接口解析租户 ID：优先 X-Tenant-ID header；其次 query ?tenant_id；最后 0（公开）
func resolveTenantID(c *gin.Context) uint {
	if v := c.GetHeader("X-Tenant-ID"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(id)
		}
	}
	if v := c.Query("tenant_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(id)
		}
	}
	return 0
}

func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

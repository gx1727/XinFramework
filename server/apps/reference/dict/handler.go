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

// Resolve 合并字典（业务最终消费入口）
// GET /api/v1/dicts/resolve?code=user_status
func (h *Handler) Resolve(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	if tenantID == 0 {
		resp.HandleError(c, ErrDictInvisible)
		return
	}
	code := c.Query("code")
	if code == "" {
		resp.BadRequest(c, "code 不能为空")
		return
	}
	rd, err := h.svc.ResolveForTenant(c.Request.Context(), tenantID, code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, resolveResponse{Dict: *rd})
}

// ResolveBatch 批量合并字典
// POST /api/v1/dicts/resolve  body: {"codes":["a","b"]}
func (h *Handler) ResolveBatch(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	if tenantID == 0 {
		resp.HandleError(c, ErrDictInvisible)
		return
	}
	var req struct {
		Codes []string `json:"codes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	out := map[string]*ResolvedDict{}
	for _, code := range req.Codes {
		rd, err := h.svc.ResolveForTenant(c.Request.Context(), tenantID, code)
		if err != nil {
			continue
		}
		out[code] = rd
	}
	resp.Success(c, gin.H{"dicts": out})
}

// ============ super_admin：平台字典 CRUD ============

// ListPlatformDicts 平台字典列表（仅 super_admin）
func (h *Handler) ListPlatformDicts(c *gin.Context) {
	var req listRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	list, total, err := h.svc.ListPlatformDicts(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, listResponse{List: list, Total: total, Page: req.Page, Size: req.Size})
}

// CreatePlatformDict 创建平台字典（仅 super_admin）
func (h *Handler) CreatePlatformDict(c *gin.Context) {
	var req platformDictCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	d, err := h.svc.CreatePlatformDict(c.Request.Context(), req)
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

// GetPlatformDict 获取平台字典
func (h *Handler) GetPlatformDict(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	d, err := h.svc.GetPlatformDict(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// UpdatePlatformDict 更新平台字典
func (h *Handler) UpdatePlatformDict(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	var req platformDictUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	d, err := h.svc.UpdatePlatformDict(c.Request.Context(), id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, d)
}

// DeletePlatformDict 删除平台字典
func (h *Handler) DeletePlatformDict(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	if err := h.svc.DeletePlatformDict(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ============ super_admin：平台字典项 CRUD ============

// ListPlatformItems 平台字典项
func (h *Handler) ListPlatformItems(c *gin.Context) {
	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	items, err := h.svc.ListPlatformItems(c.Request.Context(), dictID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"list": items, "total": int64(len(items))})
}

// CreatePlatformItem 新增平台字典项
func (h *Handler) CreatePlatformItem(c *gin.Context) {
	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	var req platformItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	item, err := h.svc.CreatePlatformItem(c.Request.Context(), dictID, req)
	if err != nil {
		if errors.Is(err, ErrDictItemCodeExists) {
			resp.Error(c, 409, "字典项编码已存在")
			return
		}
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// UpdatePlatformItem 更新平台字典项
func (h *Handler) UpdatePlatformItem(c *gin.Context) {
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
	if err := h.svc.UpdatePlatformItem(c.Request.Context(), itemID, req); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// DeletePlatformItem 删除平台字典项
func (h *Handler) DeletePlatformItem(c *gin.Context) {
	itemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}
	if err := h.svc.DeletePlatformItem(c.Request.Context(), itemID); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ============ super_admin：可见性配置 ============

// ListVisibility 平台字典的可见性矩阵
func (h *Handler) ListVisibility(c *gin.Context) {
	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	list, err := h.svc.ListVisibility(c.Request.Context(), dictID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, visibilityListResponse{List: list, Total: int64(len(list))})
}

// UpsertVisibility upsert 单条可见性配置
func (h *Handler) UpsertVisibility(c *gin.Context) {
	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	var req visibilityUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	v, err := h.svc.UpsertVisibility(c.Request.Context(), dictID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, v)
}

// DeleteVisibility 删除单条可见性配置
func (h *Handler) DeleteVisibility(c *gin.Context) {
	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	tenantID, err := parseUint(c.Param("tenant_id"))
	if err != nil {
		resp.BadRequest(c, "无效的租户ID")
		return
	}
	if err := h.svc.DeleteVisibility(c.Request.Context(), dictID, tenantID); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ============ 租户：覆盖字典项 ============

// UpsertOverride 租户 upsert 覆盖某个平台字典项
func (h *Handler) UpsertOverride(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	platformItemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}
	var req overrideUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	item, err := h.svc.UpsertOverride(c.Request.Context(), tenantID, dictID, platformItemID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

// DeleteOverride 取消租户覆盖
func (h *Handler) DeleteOverride(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	dictID, err := parseUint(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}
	platformItemID, err := parseUint(c.Param("item_id"))
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}
	if err := h.svc.DeleteOverride(c.Request.Context(), tenantID, dictID, platformItemID); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

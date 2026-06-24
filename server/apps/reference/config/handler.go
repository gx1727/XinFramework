// Package config 通用配置 - HTTP handler
//
// 拆分为三组：
//
//	BusinessHandler: 业务消费 / 租户自建（GET /configs/*, /configs/resolve*）
//	PlatformHandler: super_admin 平台 CRUD（/configs/platform/*）
//	PublicHandler:    公开读（GET /configs/public/*）
//
// 这样与 dict 模块的 handler 拆分对齐。
package config

import (
	"strconv"

	"github.com/gin-gonic/gin"

	xinContext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
)

// ============================================================================
// BusinessHandler — 业务消费 / 租户自建
// ============================================================================

type BusinessHandler struct {
	svc *Service
}

func NewBusinessHandler(svc *Service) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

// ListGroups 租户自建 group 列表
func (h *BusinessHandler) ListGroups(c *gin.Context) {
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	list, err := h.svc.repo.ListGroups(c.Request.Context(), tenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, list)
}

// GetGroup 租户视角查单个 group（支持 platform + tenant override 合并）
func (h *BusinessHandler) GetGroup(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	g, err := h.svc.repo.GetGroupByID(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	rc, err := h.svc.Resolve(c.Request.Context(), tenantID, g.Code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, rc)
}

// ListItemsByGroup 租户查某 group 下所有 item（含 override）
func (h *BusinessHandler) ListItemsByGroup(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	items, err := h.svc.repo.ListItemsByGroup(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	// RLS 已经隔离，无需再过滤 tenant_id
	resp.Success(c, items)
}

// Resolve 业务合并消费端点（与 dict 对齐）
func (h *BusinessHandler) Resolve(c *gin.Context) {
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	code := c.Query("code")
	if code == "" {
		resp.BadRequest(c, "missing code")
		return
	}
	rc, err := h.svc.Resolve(c.Request.Context(), tenantID, code)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, rc)
}

// ResolveBatch 批量 resolve（避免 N+1）
func (h *BusinessHandler) ResolveBatch(c *gin.Context) {
	var req struct {
		Codes []string `json:"codes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	all, err := h.svc.ResolveAll(c.Request.Context(), tenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	out := make(map[string]*ResolvedConfig, len(req.Codes))
	for _, code := range req.Codes {
		if rc, ok := all[code]; ok {
			out[code] = rc
		}
	}
	resp.Success(c, out)
}

// UpsertOverride 租户覆盖 platform item
func (h *BusinessHandler) UpsertOverride(c *gin.Context) {
	platformItemID, err := parseIDParam(c, "item_id")
	if err != nil {
		resp.BadRequest(c, "无效的 item_id 参数")
		return
	}
	var req struct {
		Value any `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	out, err := h.svc.UpsertOverride(c.Request.Context(), tenantID, platformItemID, req.Value)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

// DeleteOverride 租户删除对 platform item 的覆盖
func (h *BusinessHandler) DeleteOverride(c *gin.Context) {
	platformItemID, err := parseIDParam(c, "item_id")
	if err != nil {
		resp.BadRequest(c, "无效的 item_id 参数")
		return
	}
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		resp.Error(c, 400, "missing tenant_id")
		return
	}
	if err := h.svc.DeleteOverride(c.Request.Context(), tenantID, platformItemID); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ============================================================================
// PlatformHandler — super_admin 平台 CRUD
// ============================================================================

type PlatformHandler struct {
	svc *Service
}

func NewPlatformHandler(svc *Service) *PlatformHandler {
	return &PlatformHandler{svc: svc}
}

func (h *PlatformHandler) ListGroups(c *gin.Context) {
	list, err := h.svc.ListPlatformGroups(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, list)
}

func (h *PlatformHandler) GetGroup(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	g, err := h.svc.repo.GetGroupByID(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, g)
}

func (h *PlatformHandler) CreateGroup(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	scope := c.DefaultQuery("scope", "platform")
	g, err := h.svc.CreateGroup(c.Request.Context(), req, scope)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, g)
}

func (h *PlatformHandler) UpdateGroup(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req updateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	g, err := h.svc.UpdateGroup(c.Request.Context(), id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, g)
}

func (h *PlatformHandler) DeleteGroup(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.svc.DeleteGroup(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

func (h *PlatformHandler) ListItems(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	items, err := h.svc.ListPlatformItems(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, items)
}

func (h *PlatformHandler) CreateItem(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	item, err := h.svc.CreateItem(c.Request.Context(), id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

func (h *PlatformHandler) UpdateItem(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	item, err := h.svc.UpdateItem(c.Request.Context(), id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, item)
}

func (h *PlatformHandler) DeleteItem(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.svc.DeleteItem(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ===== Visibility =====

func (h *PlatformHandler) ListVisibility(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	list, err := h.svc.ListVisibility(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, list)
}

func (h *PlatformHandler) UpsertVisibility(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req struct {
		TenantID uint   `json:"tenant_id" binding:"required"`
		Access   string `json:"access" binding:"required,oneof=invisible readonly editable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	v, err := h.svc.UpsertVisibility(c.Request.Context(), id, req.TenantID, req.Access)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, v)
}

func (h *PlatformHandler) DeleteVisibility(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	tenantID, err := parseIDParam(c, "tenant_id")
	if err != nil {
		resp.BadRequest(c, "无效的 tenant_id 参数")
		return
	}
	if err := h.svc.DeleteVisibility(c.Request.Context(), id, tenantID); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// ============================================================================
// PublicHandler — 公开读（无需鉴权）
// ============================================================================

type PublicHandler struct {
	svc *Service
}

func NewPublicHandler(svc *Service) *PublicHandler {
	return &PublicHandler{svc: svc}
}

// GetPublic 取公开配置（is_public=TRUE 的 group 下所有 item）。
//
// tenantID 解析顺序：UserContext → X-Tenant-ID header → 0（platform scope）。
//
//   - tenantID > 0：在该租户 RLS 上下文中查 ci.tenant_id = X 的公开项（含该租户自建的 tenant-scope 公开项 + 平台级公开项如未覆盖）
//   - tenantID = 0 ：platform scope 模式（匿名场景也合法），RLS 过滤到 ci.tenant_id = 0 的 platform-scope 公开项
//
// 不强制要求传 tenant_id——公开读按设计对匿名/平台域开放。
func (h *PublicHandler) GetPublic(c *gin.Context) {
	tenantID := xinContext.New(c).GetTenantID()
	if tenantID == 0 {
		// 从 header 兜底（X-Tenant-ID 显式指定要看的租户）
		headerVal := c.GetHeader("X-Tenant-ID")
		if headerVal != "" {
			if tid, err := strconv.ParseUint(headerVal, 10, 64); err == nil {
				tenantID = uint(tid)
			}
		}
	}
	items, err := h.svc.ListPublicItems(c.Request.Context(), tenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, publicConfigResponse{
		Group:  "public",
		Values: indexByKey(items),
	})
}

func indexByKey(items []ConfigItem) map[string]any {
	out := make(map[string]any, len(items))
	for i := range items {
		out[items[i].Key] = items[i].Value
	}
	return out
}

// parseIDParam 解析 :id 类路径参数
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

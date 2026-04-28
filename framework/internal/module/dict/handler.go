package dict

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/repository"
	"gx1727.com/xin/framework/pkg/context"
	dictpkg "gx1727.com/xin/framework/pkg/dict"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	repo *repository.DictRepository
}

func NewHandler(repo *repository.DictRepository) *Handler {
	return &Handler{repo: repo}
}

type listResponse struct {
	List  []dictpkg.Dict `json:"list"`
	Total int64          `json:"total"`
}

func (h *Handler) List(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	list, total, err := h.repo.List(c.Request.Context(), tenantID)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, listResponse{List: list, Total: total})
}

func (h *Handler) Get(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	dictCode := c.Param("code")

	d, ok := dictpkg.Get(tenantID, dictCode)
	if !ok {
		resp.NotFound(c, "字典不存在")
		return
	}

	resp.Success(c, d)
}

// Create 创建字典数据
//
// 接收HTTP请求，解析JSON参数，调用仓储层创建字典，
// 刷新字典缓存并返回创建结果
func (h *Handler) Create(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()

	var req repository.DictCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	d, err := h.repo.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	dictpkg.RefreshDict(c.Request.Context(), tenantID, d.Code)
	resp.Success(c, d)
}

func (h *Handler) Update(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	var req struct {
		Name   string                 `json:"name"`
		Extend map[string]interface{} `json:"extend"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.repo.Update(c.Request.Context(), tenantID, uint(id), req.Name, req.Extend); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) Delete(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), tenantID, uint(id)); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) CreateItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	idStr := c.Param("id")
	dictID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的字典ID")
		return
	}

	var req repository.DictItemCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	item, err := h.repo.CreateItem(c.Request.Context(), tenantID, uint(dictID), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, item)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	idStr := c.Param("item_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}

	var req struct {
		Name   string                 `json:"name"`
		Sort   int                    `json:"sort"`
		Extend map[string]interface{} `json:"extend"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.repo.UpdateItem(c.Request.Context(), tenantID, uint(id), req.Name, req.Sort, req.Extend); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	ctx := context.New(c)
	tenantID := ctx.GetTenantID()
	idStr := c.Param("item_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的字典项ID")
		return
	}

	if err := h.repo.DeleteItem(c.Request.Context(), tenantID, uint(id)); err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, gin.H{"ok": true})
}

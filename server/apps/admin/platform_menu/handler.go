package platformmenu

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// 与 rbac/menu/handler.go 的关键区别：
//   - 没有 xc.GetTenantID() 调用（平台菜单不存在"当前租户"概念）
//   - 中间件 RequirePlatformRole("super_admin") 已在 routes.go group 级挂载
//   - 错误码经 MapRespError 映射到 4xxx（业务错误）

func (h *Handler) List(c *gin.Context) {
	list, total, err := h.svc.List(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Paginate(c, total, list)
}

func (h *Handler) Tree(c *gin.Context) {
	tree, err := h.svc.Tree(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, tree)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	result, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, result)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	result, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
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
		resp.HandleError(c, err)
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
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
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

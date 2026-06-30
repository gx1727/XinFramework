package sysmenu

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/resp"
	"gx1727.com/xin/framework/pkg/xincontext"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func operatorID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		if u, ok := uid.(uint); ok {
			return u
		}
	}
	return 0
}

func (h *Handler) List(c *gin.Context) {
	list, total, err := h.svc.List(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Paginate(c, total, list)
}

// Tree 返回当前平台用户可见的菜单树。
//
// 鉴权层级：
//   - 路由层 RequireAnyPlatformRole 已保证 PlatformRoles 非空（"需要平台级角色"）
//   - service 层 ListByUserRoles 按角色被分配的 sys_role_menus 取并集去重
//     （0024+ 移除旧 isSuperAdmin 全量分支；super_admin 靠 init_seed.sql 11.4b 绑定全菜单）
//
// 取值方式：xinc → xincontext.New(c)（Auth 中间件把 JWT claims 里的
// UserID / PlatformRoles 注入了 request context）。
func (h *Handler) Tree(c *gin.Context) {
	xc := xincontext.New(c)
	accountID := xc.UserID    // 平台用户 JWT UserID 即 account_id
	roles := xc.PlatformRoles // 已被中间件校验非空
	tree, err := h.svc.Tree(c.Request.Context(), accountID, roles)
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
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSysMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	out, err := h.svc.Create(c.Request.Context(), req, operatorID(c))
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	var req UpdateSysMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req, operatorID(c))
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		resp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, operatorID(c)); err != nil {
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

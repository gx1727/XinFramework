package task

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/resp"
	taskpkg "gx1727.com/xin/framework/pkg/task"
)

// Handler 是 task 模块的 HTTP handler。
//
// 所有路由都挂在 sys 域（/api/v1/sys/*），需 super_admin 角色。
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List 列出任务。
// GET /api/v1/sys/tasks?kind=&status=&tenant_id=&page=&size=
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	tenantID, _ := strconv.ParseUint(c.Query("tenant_id"), 10, 64)

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.svc.List(c.Request.Context(), ListFilter{
		Kind:     c.Query("kind"),
		Status:   c.Query("status"),
		TenantID: uint(tenantID),
		Limit:    size,
		Offset:   offset,
	})
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	dtos := make([]*TaskDTO, len(items))
	for i, t := range items {
		dtos[i] = ToDTO(t)
	}
	resp.Success(c, ListTasksResponse{
		Items: dtos,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Get 取单个任务详情。
// GET /api/v1/sys/tasks/:id
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的任务 ID")
		return
	}
	t, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, ToDTO(t))
}

// Cancel 取消任务。
// POST /api/v1/sys/tasks/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的任务 ID")
		return
	}
	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// Requeue 重新入队任务。
// POST /api/v1/sys/tasks/:id/requeue
func (h *Handler) Requeue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.BadRequest(c, "无效的任务 ID")
		return
	}
	if err := h.svc.Requeue(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// Stats 取队列统计。
// GET /api/v1/sys/tasks/stats
func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, StatsResponse{
		Pending:   stats.Pending,
		Running:   stats.Running,
		Succeeded: stats.Succeeded,
		Failed:    stats.Failed,
		Cancelled: stats.Cancelled,
		Dead:      stats.Dead,
	})
}

// Cleanup 清理历史任务。
// POST /api/v1/sys/tasks/cleanup?keep_days=7&statuses=succeeded,dead
//
// keep_days：保留最近 N 天的任务（默认 7 天），更早的删除。
// statuses：要清理的状态列表（默认 succeeded,dead）。
func (h *Handler) Cleanup(c *gin.Context) {
	keepDays := 7
	if v := c.Query("keep_days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			keepDays = n
		}
	}
	statuses := []taskpkg.Status{taskpkg.StatusSucceeded, taskpkg.StatusDead}
	if v := c.Query("statuses"); v != "" {
		statuses = nil
		for _, s := range splitCSV(v) {
			statuses = append(statuses, taskpkg.Status(s))
		}
	}
	before := time.Now().AddDate(0, 0, -keepDays)
	n, err := h.svc.Cleanup(c.Request.Context(), before, statuses)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"deleted": n, "before": before})
}

func splitCSV(s string) []string {
	var out []string
	cur := ""
	for _, c := range s {
		if c == ',' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

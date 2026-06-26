package task

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/resp"
)

// CronHandler 是 cron job 管理 API 的 HTTP handler。
type CronHandler struct {
	svc *CronService
}

// NewCronHandler 构造 CronHandler。
func NewCronHandler(svc *CronService) *CronHandler {
	return &CronHandler{svc: svc}
}

// List 列出所有 cron job。
// GET /api/v1/platform/cron-jobs?enabled_only=
func (h *CronHandler) List(c *gin.Context) {
	enabledOnly := c.Query("enabled_only") == "true"
	jobs, err := h.svc.List(c.Request.Context(), enabledOnly)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	dtos := make([]*CronJobDTO, 0, len(jobs))
	for _, j := range jobs {
		dtos = append(dtos, ToCronJobDTO(j))
	}
	resp.Success(c, gin.H{"items": dtos, "total": len(dtos)})
}

// Get 单个 cron job 详情。
// GET /api/v1/platform/cron-jobs/:name
func (h *CronHandler) Get(c *gin.Context) {
	name := c.Param("name")
	j, err := h.svc.Get(c.Request.Context(), name)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, ToCronJobDTO(j))
}

// Create 新建 cron job。
// POST /api/v1/platform/cron-jobs
func (h *CronHandler) Create(c *gin.Context) {
	var req CreateCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	j, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, ToCronJobDTO(j))
}

// Update 更新 cron job（按 name 定位）。
// PUT /api/v1/platform/cron-jobs/:name
func (h *CronHandler) Update(c *gin.Context) {
	name := c.Param("name")
	var req UpdateCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数格式错误")
		return
	}
	j, err := h.svc.Update(c.Request.Context(), name, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, ToCronJobDTO(j))
}

// Delete 删除 cron job。
// DELETE /api/v1/platform/cron-jobs/:name
func (h *CronHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Delete(c.Request.Context(), name); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// Enable 启用 cron job。
// POST /api/v1/platform/cron-jobs/:name/enable
func (h *CronHandler) Enable(c *gin.Context) {
	name := c.Param("name")
	j, err := h.svc.Enable(c.Request.Context(), name, true)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, ToCronJobDTO(j))
}

// Disable 禁用 cron job。
// POST /api/v1/platform/cron-jobs/:name/disable
func (h *CronHandler) Disable(c *gin.Context) {
	name := c.Param("name")
	j, err := h.svc.Enable(c.Request.Context(), name, false)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, ToCronJobDTO(j))
}

// Trigger 立即触发一次（不入 scheduler）。
// POST /api/v1/platform/cron-jobs/:name/trigger
func (h *CronHandler) Trigger(c *gin.Context) {
	name := c.Param("name")
	var req TriggerCronJobRequest
	// body 可为空，不强制要求
	_ = c.ShouldBindJSON(&req)

	taskID, err := h.svc.TriggerNow(c.Request.Context(), name, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, TriggerCronJobResponse{TaskID: taskID})
}

// 编译期保证：用 strconv 只是为了避免 unused import（未来若加 page 参数可复用）。
var _ = strconv.Atoi
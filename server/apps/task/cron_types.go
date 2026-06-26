package task

import (
	"time"

	taskpkg "gx1727.com/xin/framework/pkg/task"
	"gx1727.com/xin/framework/pkg/resp"
)

// CronJob 别名，方便业务代码短写。
type CronJob = taskpkg.CronJob

// CreateCronJobRequest 创建 cron job 的请求体。
type CreateCronJobRequest struct {
	Name        string `json:"name" binding:"required"`
	CronExpr    string `json:"cron_expr" binding:"required"`
	Timezone    string `json:"timezone"`
	Kind        string `json:"kind" binding:"required"`
	Payload     any    `json:"payload"`
	Enabled     *bool  `json:"enabled"`
	MissPolicy  string `json:"miss_policy"`
	Description string `json:"description"`
}

// UpdateCronJobRequest 更新 cron job 的请求体。
//
// Name 作为定位键（不允许改）；其他字段可选（nil = 不更新）。
type UpdateCronJobRequest struct {
	CronExpr    *string `json:"cron_expr"`
	Timezone    *string `json:"timezone"`
	Kind        *string `json:"kind"`
	Payload     any     `json:"payload"`
	Enabled     *bool   `json:"enabled"`
	MissPolicy  *string `json:"miss_policy"`
	Description *string `json:"description"`
}

// CronJobDTO 是给前端展示用的扁平化 cron job 对象。
type CronJobDTO struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	CronExpr      string     `json:"cron_expr"`
	Timezone      string     `json:"timezone"`
	Kind          string     `json:"kind"`
	Enabled       bool       `json:"enabled"`
	MissPolicy    string     `json:"miss_policy"`
	Description  string     `json:"description,omitempty"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	NextRunAt     time.Time  `json:"next_run_at"`
	LastStatus    string     `json:"last_status,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	TotalRuns     int64      `json:"total_runs"`
	TotalFailures int64      `json:"total_failures"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// 业务错误码段（1750-1759）：cron job 相关错误。
var (
	ErrCronJobExists  = resp.Err(1750, "cron job 名称已存在")
	ErrCronJobNotFound = resp.Err(1751, "cron job 不存在")
)

// ToCronJobDTO 把框架层 CronJob 转成 DTO（handle time 零值抹平）。
func ToCronJobDTO(j *CronJob) *CronJobDTO {
	if j == nil {
		return nil
	}
	return &CronJobDTO{
		ID:            j.ID,
		Name:          j.Name,
		CronExpr:      j.CronExpr,
		Timezone:      j.Timezone,
		Kind:          j.Kind,
		Enabled:       j.Enabled,
		MissPolicy:    j.MissPolicy,
		Description:  j.Description,
		LastRunAt:     j.LastRunAt,
		NextRunAt:     j.NextRunAt,
		LastStatus:    j.LastStatus,
		LastError:     j.LastError,
		TotalRuns:     j.TotalRuns,
		TotalFailures: j.TotalFails,
		CreatedAt:     j.CreatedAt,
		UpdatedAt:     j.UpdatedAt,
	}
}

// TriggerCronJobRequest 手动触发时的可选 payload 覆盖。
type TriggerCronJobRequest struct {
	Payload any `json:"payload"` // 可选；不传则用 cron job 自带的 payload
}

// TriggerCronJobResponse 触发结果。
type TriggerCronJobResponse struct {
	TaskID uint64 `json:"task_id"`
}
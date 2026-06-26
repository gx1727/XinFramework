package task

import (
	"time"

	taskpkg "gx1727.com/xin/framework/pkg/task"
)

// ListFilter 别名，让业务模块代码读起来更短。
type ListFilter = taskpkg.ListFilter

// ListTasksRequest 管理 API 列表查询参数。
type ListTasksRequest struct {
	Kind     string `form:"kind"`
	Status   string `form:"status"`
	TenantID uint   `form:"tenant_id"`
	Page     int    `form:"page"`
	Size     int    `form:"size"`
}

// ListTasksResponse 管理 API 列表响应。
type ListTasksResponse struct {
	Items []*TaskDTO `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// TaskDTO 是给前端展示用的扁平化任务对象。
type TaskDTO struct {
	ID           uint64     `json:"id"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	Priority     int        `json:"priority"`
	Attempts     int        `json:"attempts"`
	MaxAttempts  int        `json:"max_attempts"`
	LastError    string     `json:"last_error,omitempty"`
	WorkerID     string     `json:"worker_id,omitempty"`
	TimeoutSec   int        `json:"timeout_sec"`
	TenantID     uint       `json:"tenant_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RunAt        time.Time  `json:"run_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	HeartbeatAt  *time.Time `json:"heartbeat_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

// ToDTO 把 Task 转成 TaskDTO（抹平 COALESCE 的 epoch 零值）。
func ToDTO(t *taskpkg.Task) *TaskDTO {
	dto := &TaskDTO{
		ID:          t.ID,
		Kind:        t.Kind,
		Status:      string(t.Status),
		Priority:    t.Priority,
		Attempts:    t.Attempts,
		MaxAttempts: t.MaxAttempts,
		LastError:   t.LastError,
		WorkerID:    t.WorkerID,
		TimeoutSec:  t.TimeoutSec,
		TenantID:    t.TenantID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		RunAt:       t.RunAt,
	}
	if !t.StartedAt.IsZero() {
		t1 := t.StartedAt
		dto.StartedAt = &t1
	}
	if !t.HeartbeatAt.IsZero() {
		t2 := t.HeartbeatAt
		dto.HeartbeatAt = &t2
	}
	if !t.FinishedAt.IsZero() {
		t3 := t.FinishedAt
		dto.FinishedAt = &t3
	}
	return dto
}

// StatsResponse 队列统计响应。
type StatsResponse struct {
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Succeeded int64 `json:"succeeded"`
	Failed    int64 `json:"failed"`
	Cancelled int64 `json:"cancelled"`
	Dead      int64 `json:"dead"`
}
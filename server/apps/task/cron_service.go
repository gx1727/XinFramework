// Package task 是长时任务系统的业务模块入口：cron + 长时任务混合模块。
package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	taskpkg "gx1727.com/xin/framework/pkg/task"
)

// CronService 是 cron job 管理 API 的 service 层。
//
// 持有 taskpkg.CronStore + taskpkg.Queue：
//   - Create/Update/Delete/Enable 走 Store
//   - TriggerNow 直接调 Queue.Enqueue（不走 scheduler）
type CronService struct {
	store   taskpkg.CronStore
	queue   taskpkg.Queue
	scanSec int // 默认扫描周期（仅用于 RunOnceNow 不走 scheduler）
}

// NewCronService 构造 CronService。
//
// queue 为 nil 时 TriggerNow 返回错误；store 不能为 nil。
func NewCronService(store taskpkg.CronStore, queue taskpkg.Queue, scanIntervalSec int) *CronService {
	if scanIntervalSec <= 0 {
		scanIntervalSec = 60
	}
	return &CronService{store: store, queue: queue, scanSec: scanIntervalSec}
}

// List 列所有 cron job。
func (s *CronService) List(ctx context.Context, enabledOnly bool) ([]*taskpkg.CronJob, error) {
	if s.store == nil {
		return nil, ErrTaskBackendUnavailable
	}
	jobs, err := s.store.List(ctx, enabledOnly)
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}
	out := make([]*taskpkg.CronJob, len(jobs))
	for i := range jobs {
		j := jobs[i]
		out[i] = &j
	}
	return out, nil
}

// Get 取单个 cron job。
func (s *CronService) Get(ctx context.Context, name string) (*taskpkg.CronJob, error) {
	if s.store == nil {
		return nil, ErrTaskBackendUnavailable
	}
	j, err := s.store.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get cron job %q: %w", name, err)
	}
	if j == nil {
		return nil, ErrCronJobNotFound
	}
	return j, nil
}

// Create 创建 cron job。
func (s *CronService) Create(ctx context.Context, req CreateCronJobRequest) (*taskpkg.CronJob, error) {
	if s.store == nil {
		return nil, ErrTaskBackendUnavailable
	}

	// 默认值
	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	miss := req.MissPolicy
	if miss == "" {
		miss = taskpkg.MissPolicySkip
	}

	// 计算首次 next_run_at
	next, err := taskpkg.NextRunAt(req.CronExpr, tz, time.Now())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", taskpkg.ErrCronExprInvalid, err)
	}

	payload, err := marshalPayload(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	j := taskpkg.CronJob{
		Name:        req.Name,
		CronExpr:    req.CronExpr,
		Timezone:    tz,
		Kind:        req.Kind,
		Payload:     payload,
		Enabled:     enabled,
		MissPolicy:  miss,
		Description: req.Description,
		NextRunAt:   next,
	}
	if _, err := s.store.Create(ctx, j); err != nil {
		if errors.Is(err, taskpkg.ErrCronJobExists) {
			return nil, ErrCronJobExists
		}
		return nil, fmt.Errorf("create cron job: %w", err)
	}
	return s.Get(ctx, req.Name)
}

// Update 更新 cron job。
func (s *CronService) Update(ctx context.Context, name string, req UpdateCronJobRequest) (*taskpkg.CronJob, error) {
	if s.store == nil {
		return nil, ErrTaskBackendUnavailable
	}

	// 先取当前值
	cur, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	// 合并更新
	updated := *cur
	if req.CronExpr != nil {
		updated.CronExpr = *req.CronExpr
	}
	if req.Timezone != nil {
		updated.Timezone = *req.Timezone
	}
	if req.Kind != nil {
		updated.Kind = *req.Kind
	}
	if req.Payload != nil {
		payload, err := marshalPayload(req.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		updated.Payload = payload
	}
	if req.Enabled != nil {
		updated.Enabled = *req.Enabled
	}
	if req.MissPolicy != nil {
		updated.MissPolicy = *req.MissPolicy
	}
	if req.Description != nil {
		updated.Description = *req.Description
	}

	// cron_expr 或 timezone 改了 → 重算 next_run_at
	if req.CronExpr != nil || req.Timezone != nil {
		next, err := taskpkg.NextRunAt(updated.CronExpr, updated.Timezone, time.Now())
		if err != nil {
			return nil, fmt.Errorf("%w: %v", taskpkg.ErrCronExprInvalid, err)
		}
		updated.NextRunAt = next
	}

	if err := s.store.Update(ctx, updated); err != nil {
		return nil, fmt.Errorf("update cron job: %w", err)
	}
	return s.Get(ctx, name)
}

// Delete 删除 cron job。
func (s *CronService) Delete(ctx context.Context, name string) error {
	if s.store == nil {
		return ErrTaskBackendUnavailable
	}
	if err := s.store.Delete(ctx, name); err != nil {
		if errors.Is(err, taskpkg.ErrCronJobNotFound) {
			return ErrCronJobNotFound
		}
		return err
	}
	return nil
}

// Enable 启用/禁用 cron job。
func (s *CronService) Enable(ctx context.Context, name string, enabled bool) (*taskpkg.CronJob, error) {
	if s.store == nil {
		return nil, ErrTaskBackendUnavailable
	}
	if err := s.store.Enable(ctx, name, enabled); err != nil {
		if errors.Is(err, taskpkg.ErrCronJobNotFound) {
			return nil, ErrCronJobNotFound
		}
		return nil, err
	}
	return s.Get(ctx, name)
}

// TriggerNow 立即触发一次（不走 scheduler）。
//
// payload 可选；nil 时用 cron job 自带的 payload。
// 返回值：入队任务的 ID（用于前端跳转任务详情）。
func (s *CronService) TriggerNow(ctx context.Context, name string, req TriggerCronJobRequest) (uint64, error) {
	if s.store == nil {
		return 0, ErrTaskBackendUnavailable
	}
	if s.queue == nil {
		return 0, ErrTaskBackendUnavailable
	}
	j, err := s.Get(ctx, name)
	if err != nil {
		return 0, err
	}
	payload := j.Payload
	if req.Payload != nil {
		payload, err = marshalPayload(req.Payload)
		if err != nil {
			return 0, fmt.Errorf("marshal payload: %w", err)
		}
	}
	taskID, err := s.queue.Enqueue(ctx, j.Kind, payload,
		taskpkg.WithPriority(0),
		taskpkg.WithMaxAttempts(3),
	)
	if err != nil {
		return 0, fmt.Errorf("enqueue: %w", err)
	}
	return taskID, nil
}

// marshalPayload 把任意结构体或 map 序列化为 JSON 字节。
//
// payload == nil → 返回 []byte("{}")。
func marshalPayload(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	switch x := v.(type) {
	case []byte:
		return x, nil
	case string:
		return []byte(x), nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}
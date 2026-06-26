// Package task 是长时任务系统的业务模块入口：
//   - 注册 platform 域管理 API（list / get / cancel / requeue / stats / cleanup）
//   - 启动 worker pool（轮询 / 心跳 / 僵死回收）
//   - 内置 cleanup handler（cleanup_expired_locks / cleanup_old_tasks）
//
// 包名冲突说明：与 framework/pkg/task 同名，本包通过别名
//
//	import taskpkg "gx1727.com/xin/framework/pkg/task"
//
// 引用框架层类型。这样业务模块文件名仍叫 task / routes / handler。
package task

import (
	"context"
	"fmt"
	"time"

	taskpkg "gx1727.com/xin/framework/pkg/task"
)

// Service 是 task 业务模块的 service 层。
//
// 持有 taskpkg.Queue（来自 plugin.AppContext），管理 API 通过它操作任务。
type Service struct {
	q taskpkg.Queue
}

// NewService 构造 Service。
func NewService(q taskpkg.Queue) *Service {
	return &Service{q: q}
}

// List 分页列出任务。
func (s *Service) List(ctx context.Context, f taskpkg.ListFilter) ([]*taskpkg.Task, int, error) {
	if s.q == nil {
		return nil, 0, ErrTaskBackendUnavailable
	}
	items, total, err := s.q.List(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	return items, total, nil
}

// Get 取单个任务详情。
func (s *Service) Get(ctx context.Context, id uint64) (*taskpkg.Task, error) {
	if s.q == nil {
		return nil, ErrTaskBackendUnavailable
	}
	t, err := s.q.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get task %d: %w", id, err)
	}
	if t == nil {
		return nil, ErrTaskNotFound
	}
	return t, nil
}

// Cancel 取消任务（pending 或 running 状态可取消）。
func (s *Service) Cancel(ctx context.Context, id uint64) error {
	if s.q == nil {
		return ErrTaskBackendUnavailable
	}
	t, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if t.Status.IsTerminal() {
		return ErrTaskInvalidTransition
	}
	if err := s.q.Cancel(ctx, id); err != nil {
		return fmt.Errorf("cancel task %d: %w", id, err)
	}
	return nil
}

// Requeue 重新入队 failed/dead 任务（attempts 归零）。
func (s *Service) Requeue(ctx context.Context, id uint64) error {
	if s.q == nil {
		return ErrTaskBackendUnavailable
	}
	t, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if t.Status != taskpkg.StatusFailed && t.Status != taskpkg.StatusDead {
		return ErrTaskInvalidTransition
	}
	if err := s.q.Requeue(ctx, id); err != nil {
		return fmt.Errorf("requeue task %d: %w", id, err)
	}
	return nil
}

// Stats 取队列统计。
func (s *Service) Stats(ctx context.Context) (taskpkg.Stats, error) {
	if s.q == nil {
		return taskpkg.Stats{}, ErrTaskBackendUnavailable
	}
	stats, err := s.q.Stats(ctx)
	if err != nil {
		return taskpkg.Stats{}, fmt.Errorf("stats: %w", err)
	}
	return stats, nil
}

// Cleanup 删除 N 天前的终态任务。
func (s *Service) Cleanup(ctx context.Context, before time.Time, statuses []taskpkg.Status) (int, error) {
	if s.q == nil {
		return 0, ErrTaskBackendUnavailable
	}
	if len(statuses) == 0 {
		statuses = []taskpkg.Status{taskpkg.StatusSucceeded}
	}
	n, err := s.q.Cleanup(ctx, before, statuses)
	if err != nil {
		return 0, fmt.Errorf("cleanup: %w", err)
	}
	return n, nil
}

// Enqueue 业务侧入队（供同进程其他模块调用，非 HTTP API）。
func (s *Service) Enqueue(ctx context.Context, kind string, payload []byte, opts ...taskpkg.EnqueueOption) (uint64, error) {
	if s.q == nil {
		return 0, ErrTaskBackendUnavailable
	}
	id, err := s.q.Enqueue(ctx, kind, payload, opts...)
	if err != nil {
		return 0, fmt.Errorf("enqueue %s: %w", kind, err)
	}
	return id, nil
}
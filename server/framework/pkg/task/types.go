// Package task 提供长时任务系统：异步任务队列、worker 池、重试、DLQ。
//
// 核心组件：
//   - Queue：入队 / 抢占（SKIP LOCKED）/ 完成 / 失败 / 心跳 / 僵死回收
//   - Worker：轮询 + 执行 + 重试 + 心跳 + 优雅停机
//   - Handler：业务侧实现，每个 kind 一个
//   - Registry：handler 注册表
//
// 存储：用 PostgreSQL 的 background_tasks 表（migrations/background_tasks.sql）。
// 不引入 Redis Stream / asynq 等额外依赖——MVP 阶段 PG 完全够用。
//
// 调用方：
//   - login_security.Notifier：把发短信/邮件改为入队
//   - apps/tenant/message.Service.Send：消息入库后入队推送任务
//   - apps/flag.Handler：合成头像框（耗时）改为入队
//   - apps/task/module.go：内置清理任务（cleanup_expired_locks / cleanup_old_tasks）
package task

import (
	"encoding/json"
	"time"
)

// Status 任务状态枚举。
type Status string

const (
	StatusPending   Status = "pending"   // 待执行
	StatusRunning   Status = "running"   // 正在执行
	StatusSucceeded Status = "succeeded" // 已成功
	StatusFailed    Status = "failed"    // 已失败（attempts < max_attempts 时可重试，最终状态仍可能回 pending）
	StatusCancelled Status = "cancelled" // 手动取消
	StatusDead      Status = "dead"      // 超过 max_attempts 仍失败，进死信队列
)

// IsTerminal 报告状态是否为终态（不再变化）。
//
// 终态：succeeded / failed / cancelled / dead。
// pending / running 都不是终态。
func (s Status) IsTerminal() bool {
	switch s {
	case StatusSucceeded, StatusFailed, StatusCancelled, StatusDead:
		return true
	}
	return false
}

// Task 一条任务的完整记录。
//
// 注意：Payload 是 []byte（JSONB 反序列化结果），handler 自行决定如何 Unmarshal。
// 这样可以避免 task 包依赖所有业务 payload 类型。
type Task struct {
	ID           uint64
	Kind         string
	Payload      []byte
	Status       Status
	Priority     int
	RunAt        time.Time
	Attempts     int
	MaxAttempts  int
	LastError    string
	WorkerID     string
	StartedAt    time.Time
	HeartbeatAt  time.Time
	FinishedAt   time.Time
	TimeoutSec   int
	TenantID     uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// EnqueueOption 入队时的可选项。
type EnqueueOption func(*enqueueConfig)

type enqueueConfig struct {
	priority     int
	runAt        time.Time
	maxAttempts  int
	timeoutSec   int
	tenantID     uint
	kind         string
	payload      []byte
}

// WithPriority 设置任务优先级（默认 0，数字越大越优先）。
func WithPriority(p int) EnqueueOption {
	return func(c *enqueueConfig) { c.priority = p }
}

// WithRunAt 设置首次执行时间（用于延迟任务）。
//
// 未设置时默认 NOW()，即立刻可执行。
func WithRunAt(t time.Time) EnqueueOption {
	return func(c *enqueueConfig) { c.runAt = t }
}

// WithMaxAttempts 设置最大重试次数（默认 3）。
func WithMaxAttempts(n int) EnqueueOption {
	return func(c *enqueueConfig) { c.maxAttempts = n }
}

// WithTimeout 设置单次执行超时秒数（默认 300）。
func WithTimeout(sec int) EnqueueOption {
	return func(c *enqueueConfig) { c.timeoutSec = sec }
}

// WithTenantID 关联租户 ID（管理 API 按 tenant_id 分页时使用）。
func WithTenantID(tid uint) EnqueueOption {
	return func(c *enqueueConfig) { c.tenantID = tid }
}

// MarshalPayload 把任意结构体序列化为 JSON 字节，便于 Enqueue 时直接传入。
//
// 推荐用法：queue.Enqueue(ctx, "send_email", task.MarshalPayload(myPayload))
func MarshalPayload(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

// WorkerStats worker 当前运行状态的快照。
type WorkerStats struct {
	WorkerID        string
	Started         bool
	ProcessedTotal  uint64 // 累计成功数
	FailedTotal     uint64 // 累计失败数（不含进 DLQ 的）
	DeadTotal       uint64 // 累计进 DLQ 数
	InFlight        int    // 当前正在执行的任务数
}
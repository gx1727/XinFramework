package task

import (
	"context"
	"errors"
	"sync"
)

// ErrHandlerNotFound 业务侧 Enqueue 时 kind 未注册到 Registry。
//
// 调用方应立即返回这个错误（不要重试，因为重试也没用）。
var ErrHandlerNotFound = errors.New("task: handler not registered for kind")

// ErrHandlerAlreadyExists 重复注册同名 kind。
var ErrHandlerAlreadyExists = errors.New("task: handler already registered for kind")

// Handler 业务侧任务处理器。
//
// 每个 kind 对应一个 Handler。handler.Handle 返回 error 时，
// Worker 会按 retry 策略决定重试或进 DLQ；返回 nil 时记 succeeded。
//
// 实现者职责：
//   - 必须考虑幂等性（任务可能被多次执行，例如 worker 僵死后被回收）
//   - 严禁在 Handle 内做长事务（会占用 worker slot 太久）
//   - 严禁在 Handle 内 panic（worker 捕获 panic 转成 error）
type Handler interface {
	Kind() string
	Handle(ctx context.Context, t *Task) error
	// Timeout 是单次执行超时；返回 0 时 worker 用 queue 默认值。
	Timeout() int
}

// HandlerFunc 是 Handler 的函数式适配器（避免简单任务写 struct）。
//
// 用法：task.RegisterHandler("send_email", task.HandlerFunc(func(ctx, t) error { ... }))
type HandlerFunc struct {
	KindStr  string
	HandleFn func(ctx context.Context, t *Task) error
	TimeoutV int
}

// Kind 实现 Handler。
func (h HandlerFunc) Kind() string { return h.KindStr }

// Handle 实现 Handler。
func (h HandlerFunc) Handle(ctx context.Context, t *Task) error { return h.HandleFn(ctx, t) }

// Timeout 实现 Handler。
func (h HandlerFunc) Timeout() int { return h.TimeoutV }

// Registry 全局 Handler 注册表。
//
// 进程级单例（var defaultRegistry HandlerRegistry），与 task 包其他单例一致。
// 注册动作一般在 init() / module.Register 阶段完成。
type Registry interface {
	Register(h Handler) error
	Get(kind string) (Handler, bool)
	Names() []string
}

type defaultRegistry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

// NewRegistry 返回新的 Registry（用于测试隔离或私有场景）。
func NewRegistry() Registry {
	return &defaultRegistry{handlers: make(map[string]Handler)}
}

var (
	globalRegistryMu sync.RWMutex
	globalRegistry   = NewRegistry()
)

// DefaultRegistry 返回进程级 Registry。
//
// 模块启动时调 RegisterHandler 注册 handler，worker 内部使用 DefaultRegistry 查找。
func DefaultRegistry() Registry {
	globalRegistryMu.RLock()
	defer globalRegistryMu.RUnlock()
	return globalRegistry
}

// RegisterHandler 注册 handler 到全局 Registry。
//
// 重复注册同名 kind 返回 ErrHandlerAlreadyExists（启动期 fail-fast）。
func RegisterHandler(h Handler) error {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()

	if _, ok := globalRegistry.(Registry); !ok {
		return errors.New("task: global registry is broken")
	}
	r := globalRegistry.(*defaultRegistry)
	if _, exists := r.handlers[h.Kind()]; exists {
		return ErrHandlerAlreadyExists
	}
	r.handlers[h.Kind()] = h
	return nil
}

// Register 是 Registry 接口实现（支持私有 registry 场景）。
func (r *defaultRegistry) Register(h Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[h.Kind()]; exists {
		return ErrHandlerAlreadyExists
	}
	r.handlers[h.Kind()] = h
	return nil
}

// Get 是 Registry 接口实现。
func (r *defaultRegistry) Get(kind string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[kind]
	return h, ok
}

// Names 是 Registry 接口实现。
func (r *defaultRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.handlers))
	for k := range r.handlers {
		out = append(out, k)
	}
	return out
}

// LookupHandler 是 worker 用的便捷函数：从全局 Registry 查 handler。
//
// 未找到返回 (nil, ErrHandlerNotFound)。
func LookupHandler(kind string) (Handler, error) {
	r := DefaultRegistry()
	h, ok := r.Get(kind)
	if !ok {
		return nil, ErrHandlerNotFound
	}
	return h, nil
}
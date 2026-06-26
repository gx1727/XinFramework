package task

import (
	"math"
	"time"
)

// BackoffStrategy 重试退避策略枚举。
type BackoffStrategy string

const (
	BackoffExponential BackoffStrategy = "exponential" // 2^n × initial_delay
	BackoffLinear      BackoffStrategy = "linear"      // n × initial_delay
	BackoffFixed       BackoffStrategy = "fixed"       // 始终 initial_delay
)

// BackoffConfig 退避参数。
type BackoffConfig struct {
	Strategy        BackoffStrategy
	InitialDelay    time.Duration // 默认 30s
	MaxDelay        time.Duration // 默认 1h
}

// DefaultBackoff 返回推荐默认配置：指数退避，30s → 1h。
func DefaultBackoff() BackoffConfig {
	return BackoffConfig{
		Strategy:     BackoffExponential,
		InitialDelay: 30 * time.Second,
		MaxDelay:     1 * time.Hour,
	}
}

// NextDelay 计算第 attempts 次失败后，下次重试的等待时长。
//
// attempts 是当前已失败次数（从 1 开始），即调用 NextDelay(1) 表示
// "失败 1 次后下次重试要等多久"。
//
// 返回值会被 MaxDelay 截断。
func (c BackoffConfig) NextDelay(attempts int) time.Duration {
	if attempts <= 0 {
		attempts = 1
	}
	initial := c.InitialDelay
	if initial <= 0 {
		initial = 30 * time.Second
	}
	maxDelay := c.MaxDelay
	if maxDelay <= 0 {
		maxDelay = 1 * time.Hour
	}
	var d time.Duration
	switch c.Strategy {
	case BackoffLinear:
		d = initial * time.Duration(attempts)
	case BackoffFixed:
		d = initial
	default: // exponential
		d = initial * time.Duration(math.Pow(2, float64(attempts-1)))
	}
	if d > maxDelay {
		d = maxDelay
	}
	if d < 0 {
		d = initial
	}
	return d
}
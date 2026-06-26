package task

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

// 管理 API 错误码段（1700-1799）。
//
// 错误语义：
//   - 1700 = 任务不存在
//   - 1701 = 当前状态不允许该操作（如对 succeeded 任务调 cancel）
//   - 1702 = 后端不可用
var (
	ErrTaskNotFound          = resp.Err(1700, "任务不存在")
	ErrTaskInvalidTransition = resp.Err(1701, "当前任务状态不允许此操作")
	ErrTaskBackendUnavailable = resp.Err(1702, "任务队列后端不可用")
)

var errTaskInvalidTransition = errors.New("task: invalid state transition")
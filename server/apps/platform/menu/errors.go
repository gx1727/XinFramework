package menu

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

// 平台菜单错误码段：15001-15999
// 见 framework/pkg/resp/errors.go 的分段约定
const (
	CodePlatformMenu = 15000
)

var (
	ErrMenuNotFound       = resp.Err(15001, "平台菜单不存在")
	ErrMenuCodeExists     = resp.Err(15002, "平台菜单编码已存在")
	ErrBackendUnavailable = resp.Err(15003, "服务后端未初始化或不可用")
)

// 内部错误（不会直接返回前端）
var (
	errMenuNotFoundDB = errors.New("platform menu not found in db")
)

package menu

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrMenuNotFound       = resp.Err(5001, "菜单不存在")
	ErrMenuCodeExists     = resp.Err(5002, "菜单编码已存在")
	ErrBackendUnavailable = resp.Err(5003, "服务后端未初始化或不可用")
)

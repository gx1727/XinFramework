package resource

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrResourceNotFound     = resp.Err(8001, "资源不存在")
	ErrResourceCodeExists   = resp.Err(8002, "资源编码已存在")
	ErrCannotDeleteResource = resp.Err(8003, "不能删除系统资源")
	ErrBackendUnavailable   = resp.Err(8004, "服务后端未初始化或不可用")
)

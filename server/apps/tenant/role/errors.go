package role

import (
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrRoleNotFound       = resp.Err(4001, "角色不存在")
	ErrRoleCodeExists     = resp.Err(4002, "角色编码已存在")
	ErrCannotDeleteAdmin  = resp.Err(4003, "不能删除管理员角色")
	ErrBackendUnavailable = resp.Err(4004, "服务后端未初始化或不可用")
)

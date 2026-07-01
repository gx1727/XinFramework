package sysrole

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

const (
	CodeSysRole = 15200
)

var (
	ErrSysRoleNotFound      = resp.Err(15201, "平台角色不存在")
	ErrSysRoleCodeExists    = resp.Err(15202, "平台角色编码已存在")
	ErrSysRoleInvalidStatus = resp.Err(15203, "状态值非法")
	ErrBackendUnavailable   = resp.Err(15299, "服务后端未初始化或不可用")
)

var (
	errSysRoleNotFoundDB = errors.New("sys_role not found in db")
)

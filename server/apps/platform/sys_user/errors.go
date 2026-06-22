package sysuser

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

// 平台用户错误码段：15100-15199
const (
	CodeSysUser = 15100
)

var (
	ErrSysUserNotFound        = resp.Err(15101, "平台用户不存在")
	ErrSysUserCodeExists      = resp.Err(15102, "平台用户编码已存在")
	ErrSysUserAccountRequired = resp.Err(15103, "账号 ID 必填")
	ErrSysUserAccountMissing  = resp.Err(15104, "账号不存在")
	ErrSysUserAlreadyExists   = resp.Err(15105, "该账号已绑定平台用户")
	ErrSysUserInvalidStatus   = resp.Err(15106, "状态值非法")
	ErrBackendUnavailable     = resp.Err(15199, "服务后端未初始化或不可用")
)

// 内部错误（不会直接返回前端）
var (
	errSysUserNotFoundDB = errors.New("sys_user not found in db")
)

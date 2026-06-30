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
	// 15107-15109：创建可登录平台用户（一并建 accounts 行）专属错误
	ErrSysUserPhoneRequired      = resp.Err(15107, "新建账号时手机号必填")
	ErrSysUserPasswordInvalid    = resp.Err(15108, "密码长度需在 6-32 之间")
	ErrSysUserPhoneExists        = resp.Err(15109, "该手机号已注册账号")
	ErrSysUserEmailExists        = resp.Err(15110, "该邮箱已注册账号")
	ErrSysUserAccountRepoMissing = resp.Err(15111, "账号仓储未初始化")
	ErrBackendUnavailable        = resp.Err(15199, "服务后端未初始化或不可用")
)

// 内部错误（不会直接返回前端）
var (
	errSysUserNotFoundDB = errors.New("sys_user not found in db")
)

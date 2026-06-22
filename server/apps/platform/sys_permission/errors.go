package syspermission

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

const (
	CodeSysPermission = 15400
)

var (
	ErrSysPermissionNotFound    = resp.Err(15401, "平台权限码不存在")
	ErrSysPermissionCodeExists  = resp.Err(15402, "平台权限码已存在")
	ErrSysPermissionInvalidCode = resp.Err(15403, "权限码格式非法")
	ErrBackendUnavailable       = resp.Err(15499, "服务后端未初始化或不可用")
)

var (
	errSysPermissionNotFoundDB = errors.New("sys_permission not found in db")
)

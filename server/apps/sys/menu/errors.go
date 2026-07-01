package sysmenu

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

const (
	CodeSysMenu = 15300
)

var (
	ErrSysMenuNotFound    = resp.Err(15301, "平台菜单不存在")
	ErrSysMenuCodeExists  = resp.Err(15302, "平台菜单编码已存在")
	ErrBackendUnavailable = resp.Err(15399, "服务后端未初始化或不可用")
)

var (
	errSysMenuNotFoundDB = errors.New("sys_menu not found in db")
)

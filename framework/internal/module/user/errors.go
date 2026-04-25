package user

import (
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrUserNotFound = resp.NewError(2001, "用户不存在")
)

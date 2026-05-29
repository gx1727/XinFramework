package user

import (
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrUserNotFound = resp.Err(2001, "用户不存在")
)

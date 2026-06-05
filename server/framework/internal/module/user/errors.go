package user

import (
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrUserNotFound = resp.Err(2001, "用户不存在")

	// ErrOrgNotFound 主组织 ID 找不到（同租户内）时返回
	ErrOrgNotFound = resp.Err(2010, "组织不存在或不属于当前租户")
)

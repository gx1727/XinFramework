package flag

import "gx1727.com/xin/framework/pkg/resp"

// Flag 模块业务错误码定义
var (
	ErrFrameNotFound  = resp.NewError(15001, "头像框不存在")
	ErrSpaceNotFound  = resp.NewError(15002, "活动空间不存在")
	ErrGenerateFailed = resp.NewError(15003, "头像生成失败")
	ErrAvatarNotFound = resp.NewError(15004, "头像不存在")
)

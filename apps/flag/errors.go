package flag

import "gx1727.com/xin/framework/pkg/resp"

// Flag 模块业务错误码定义
var (
	ErrFrameNotFound        = resp.Err(13001, "头像框不存在")
	ErrSpaceNotFound        = resp.Err(13002, "活动空间不存在")
	ErrGenerateFailed       = resp.Err(13003, "头像生成失败")
	ErrAvatarNotFound       = resp.Err(13004, "头像不存在")
	ErrCategoryCodeExists   = resp.Err(13005, "相框分类编码已存在")
	ErrCreateCategoryFailed = resp.Err(13006, "创建相框分类失败")
	ErrUpdateCategoryFailed = resp.Err(13007, "更新相框分类失败")
)

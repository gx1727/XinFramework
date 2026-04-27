package asset

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrFileTooLarge    = resp.NewError(14001, "文件大小超出限制")
	ErrUnsupportedType = resp.NewError(14002, "不支持的文件类型")
	ErrUploadFailed    = resp.NewError(14003, "文件上传失败")
	ErrFileNotFound    = resp.NewError(14004, "文件不存在")
)

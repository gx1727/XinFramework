package asset

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrFileTooLarge    = resp.Err(9001, "文件大小超出限制")
	ErrUnsupportedType = resp.Err(9002, "不支持的文件类型")
	ErrUploadFailed    = resp.Err(9003, "文件上传失败")
	ErrFileNotFound    = resp.Err(9004, "文件不存在")
)

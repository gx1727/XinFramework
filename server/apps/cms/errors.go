package cms

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

// Cms 模块错误码段：14xxx（见 resp.CodeCMS）。
// 示例模块，没有 service 层，handler 直接调 mapRepoError 翻译。
var (
	ErrPostNotFound = resp.Err(resp.CodeCMS+1, "文章不存在")
)

// DB 层 sentinel 错误（仅在 repository 实现里返回）。
var (
	ErrPostNotFoundDB = errors.New("post not found")
)

// mapRepoError 把 DB 层 sentinel 翻译为 BizError，未识别 error 原样返回。
func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrPostNotFoundDB) {
		return ErrPostNotFound
	}
	return err
}
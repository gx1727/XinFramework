package resource

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrResourceNotFound     = resp.Err(8001, "资源不存在")
	ErrResourceCodeExists   = resp.Err(8002, "资源编码已存在")
	ErrCannotDeleteResource = resp.Err(8003, "不能删除系统资源")
	ErrBackendUnavailable   = resp.Err(8004, "服务后端未初始化或不可用")
	// ErrResourceInvalidCode 权限码格式错误：必须为 resource:action 或 resource:*（仅含一个冒号）。
	// 与 apps/sys/permission/service.go permissionCodeValid 规则一致。
	ErrResourceInvalidCode = resp.Err(8005, "权限码格式错误，必须为 resource:action 或 resource:*（仅含一个冒号）")
)

// mapRepoError 把 DB 层 sentinel 翻译为 BizError，未识别 error 原样返回。
// service 层所有对 resourceRepo 的调用结果都要过这一层。
func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrResourceNotFoundDB) {
		return ErrResourceNotFound
	}
	if errors.Is(err, ErrResourceCodeExistsDB) {
		return ErrResourceCodeExists
	}
	return err
}

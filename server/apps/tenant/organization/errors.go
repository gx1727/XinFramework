package organization

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrOrgNotFound        = resp.Err(6001, "组织不存在")
	ErrOrgCodeExists      = resp.Err(6002, "组织编码已存在")
	ErrCannotDeleteRoot   = resp.Err(6003, "不能删除根组织")
	ErrBackendUnavailable = resp.Err(6004, "服务后端未初始化或不可用")

	// ErrOrgHasUsers 组织下还有未删用户，禁止删除（直接 + 后代）
	ErrOrgHasUsers = resp.Err(6005, "组织下仍有用户，无法删除")
)

// mapRepoError 把 DB 层 sentinel 翻译为 BizError，未识别 error 原样返回。
// service 层所有对 orgRepo 的调用结果都要过这一层，
// 否则 handler 拿到的就是裸 sentinel，被 HandleError 当成未知错误 → 500。
func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrOrgNotFoundDB) {
		return ErrOrgNotFound
	}
	return err
}

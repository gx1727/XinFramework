package organization

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrOrgNotFound        = resp.Err(6001, "组织不存在")
	ErrOrgCodeExists      = resp.Err(6002, "组织编码已存在")
	ErrCannotDeleteRoot   = resp.Err(6003, "不能删除根组织")
	ErrBackendUnavailable = resp.Err(6004, "服务后端未初始化或不可用")
)

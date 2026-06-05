package tenant

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrTenantNotFound     = resp.Err(3001, "租户不存在")
	ErrTenantCodeExists   = resp.Err(3002, "租户编码已存在")
	ErrTenantDisabled     = resp.Err(3003, "租户已被禁用")
	ErrTenantCreateFailed = resp.Err(3004, "创建租户失败")
	ErrTenantUpdateFailed = resp.Err(3005, "更新租户失败")
	ErrTenantDeleteFailed = resp.Err(3006, "删除租户失败")
	ErrTenantListFailed   = resp.Err(3007, "查询租户列表失败")
	ErrBackendUnavailable = resp.Err(3008, "服务后端未初始化或不可用")
)

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrTenantNotFoundDB):
		return ErrTenantNotFound
	case errors.Is(err, ErrTenantCodeExistsDB):
		return ErrTenantCodeExists
	default:
		return err
	}
}

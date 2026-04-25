package tenant

import (
	"errors"

	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrTenantNotFound     = resp.NewError(2001, "租户不存在")
	ErrTenantCodeExists   = resp.NewError(2002, "租户编码已存在")
	ErrTenantDisabled     = resp.NewError(2003, "租户已被禁用")
	ErrTenantCreateFailed = resp.NewError(2004, "创建租户失败")
	ErrTenantUpdateFailed = resp.NewError(2005, "更新租户失败")
	ErrTenantDeleteFailed = resp.NewError(2006, "删除租户失败")
	ErrTenantListFailed   = resp.NewError(2007, "查询租户列表失败")
	ErrBackendUnavailable = resp.NewError(2008, "服务后端未初始化或不可用")
)

func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, model.ErrTenantNotFound):
		return ErrTenantNotFound
	case errors.Is(err, model.ErrTenantCodeExists):
		return ErrTenantCodeExists
	default:
		return err
	}
}

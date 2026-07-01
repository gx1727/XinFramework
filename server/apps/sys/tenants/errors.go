package tenants

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrTenantNotFound           = resp.Err(3001, "租户不存在")
	ErrTenantCodeExists         = resp.Err(3002, "租户编码已存在")
	ErrTenantDisabled           = resp.Err(3003, "租户已被禁用")
	ErrTenantCreateFailed       = resp.Err(3004, "创建租户失败")
	ErrTenantUpdateFailed       = resp.Err(3005, "更新租户失败")
	ErrTenantDeleteFailed       = resp.Err(3006, "删除租户失败")
	ErrTenantListFailed         = resp.Err(3007, "查询租户列表失败")
	ErrBackendUnavailable       = resp.Err(3008, "服务后端未初始化或不可用")
	ErrTenantHasUsers           = resp.Err(3009, "租户下存在未删除用户，禁止删除")
	ErrTenantPurgeNotAllowed    = resp.Err(3010, "租户未软删，禁止硬删；请先调 DELETE /tenants/:id")
	ErrTenantPurgeFailed        = resp.Err(3011, "硬删租户失败")
	ErrTenantImpersonateNoAdmin = resp.Err(3012, "租户尚未配置管理员账号，无法模拟登录")
	ErrTenantImpersonateFailed  = resp.Err(3013, "模拟登录失败")
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
	case errors.Is(err, ErrNoAdminUserDB):
		return ErrTenantImpersonateNoAdmin
	default:
		return err
	}
}

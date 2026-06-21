package platformtenant

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
	// ErrTenantHasUsers 删除前置校验：租户下还存在未删除的用户，禁止软删。
	// 业务要求：必须先把所有用户迁出 / 软删，才能走 tenant.Delete，避免留下幽灵租户。
	ErrTenantHasUsers = resp.Err(3009, "租户下存在未删除用户，禁止删除")
	// ErrTenantPurgeNotAllowed 硬删前置校验：租户未被软删，禁止 purge。
	// 流程要求：必须先 Delete（软删），走完审计 + 用户清理，再 Purge（硬删）。
	ErrTenantPurgeNotAllowed = resp.Err(3010, "租户未软删，禁止硬删；请先调用 DELETE /tenants/:id")
	// ErrTenantPurgeFailed 硬删过程失败（如 FK 约束、孤儿数据），事务回滚后留软删状态供排查。
	ErrTenantPurgeFailed = resp.Err(3011, "硬删租户失败")
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

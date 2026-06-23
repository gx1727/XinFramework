package tenant

import (
	"context"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

// Module returns the platform_tenant module as a BaseModule.
//
//
// 模块名约定："platform_tenant"（与 apps/tenant/menu / apps/platform/menu 一致）。
// 在 cfg.Module: 里以 "platform_tenant" 标识，且属于 alwaysOn 列表（启动必需）。
//
// 本模块的双重职责：
//  1. 提供 super_admin 用的租户 CRUD API（路由挂 /admin/platform-tenants）
//  2. 把自己的 TenantRepository 写进 AppContext.Writer，
//     让 framework 的 extapi.Provider、其他业务模块跨模块读取租户数据
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "platform_tenant",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB
			w.SetTenantRepo(&tenantPkgAdapter{repo: NewTenantRepository(pool).(*PostgresTenantRepository)})
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			h := NewHandler(NewService(pool, NewTenantRepository(pool)))
			Register(protected, h)
		},
	}
}

// tenantPkgAdapter 把 *PostgresTenantRepository 适配成 pkg/tenant.TenantRepository。
// GetTenantRecord 已经在 *PostgresTenantRepository 上实现了字段 copy（见 repository.go），
// adapter 只是 1 行 forwarder——把 ctx 上的 GetByID 调用转发到 repo 的 GetTenantRecord。
type tenantPkgAdapter struct {
	repo *PostgresTenantRepository
}

func (a *tenantPkgAdapter) GetByID(ctx context.Context, id uint) (*pkgtenant.TenantRecord, error) {
	return a.repo.GetTenantRecord(ctx, id)
}

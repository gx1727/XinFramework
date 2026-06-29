package tenants

import (
	"context"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

// Module returns the platform_tenant module as a BaseModule.
//
//
// 模块名约定："platform_tenant"（管 tenants 表的 CRUD；与 sys_menu 同属 platform 域）　
// 圀cfg.Module: 里以 "platform_tenant" 标识，且属于 alwaysOn 列表（启动必需）　
//
// 本模块的双重职责：
//  1. 提供 super_admin 用的租户 CRUD API（路由挂 /admin/platform-tenants：
//  2. 把自己的 TenantRepository 写进 AppContext.Writer：
//     讀framework 皀extapi.Provider、其他业务模块跨模块读取租户数据
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "platform_tenant",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB.Raw()
			w.SetTenantRepo(&tenantPkgAdapter{repo: NewTenantRepository(pool).(*PostgresTenantRepository)})
			return nil
		},
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			protected := slots.MustGet(plugin.SlotProtected).Group
			pool := app.DB.Raw()
			h := NewHandler(NewService(pool, NewTenantRepository(pool)))
			Register(protected, h)
		},
	}
}

// tenantPkgAdapter 技*PostgresTenantRepository 适配戀pkg/tenant.TenantRepository　
// GetTenantRecord 已经圀*PostgresTenantRepository 上实现了字段 copy（见 repository.go），
// adapter 只是 1 血forwarder——把 ctx 上的 GetByID 调用转发刀repo 皀GetTenantRecord　
type tenantPkgAdapter struct {
	repo *PostgresTenantRepository
}

func (a *tenantPkgAdapter) GetByID(ctx context.Context, id uint) (*pkgtenant.TenantRecord, error) {
	return a.repo.GetTenantRecord(ctx, id)
}

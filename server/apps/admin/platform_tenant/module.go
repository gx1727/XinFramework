package platformtenant

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
// 模块名约定："platform_tenant"（与 apps/rbac/menu / apps/admin/platform_menu 一致）。
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
			w.SetTenantRepo(&tenantPkgAdapter{repo: NewTenantRepository(pool)})
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			h := NewHandler(NewService(pool, NewTenantRepository(pool)))
			Register(protected, h)
		},
	}
}

// tenantPkgAdapter wraps apps/admin/platform_tenant's TenantRepository so it
// satisfies pkg/tenant.TenantRepository (returns *pkg/tenant.TenantRecord).
//
// 历史背景：本模块从 apps/boot/tenant 重命名而来，原本就有这个 adapter。
// 消费者（cms、extapi）的"窄接口"，platform_tenant 通过 adapter 提供实现。
type tenantPkgAdapter struct {
	repo TenantRepository
}

func (a *tenantPkgAdapter) GetByID(ctx context.Context, id uint) (*pkgtenant.TenantRecord, error) {
	t, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pkgtenant.TenantRecord{
		ID: t.ID, Code: t.Code, Name: t.Name, Status: t.Status,
		Contact: t.Contact, Phone: t.Phone, Email: t.Email,
		Province: t.Province, City: t.City, Area: t.Area, Address: t.Address,
		Config: t.Config, Dashboard: t.Dashboard,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}, nil
}

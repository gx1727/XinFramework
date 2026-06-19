package tenant

import (
	"context"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/bootx"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the tenant module as a BaseModule.
//
// Phase 4 changes: db.Get() → bootx.Pool()（过渡期全局变量已删除）
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "tenant",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := bootx.Pool()
			w.SetTenantRepo(&tenantPkgAdapter{repo: NewTenantRepository(pool)})
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			h := NewHandler(NewService(pool, NewTenantRepository(pool)))
			Register(protected, h)
		},
	}
}

// tenantPkgAdapter wraps apps/boot/tenant's TenantRepository so it
// satisfies pkg/tenant.TenantRepository (returns *pkg/tenant.TenantRecord).
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

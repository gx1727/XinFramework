package tenant

import (
	"context"

	"github.com/gin-gonic/gin"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())

	// Phase 2: register this module's TenantRepository with the
	// framework's public pkg/tenant registry so that framework's
	// extapi.Provider (in framework/internal) can resolve tenant
	// data without importing apps/.
	//
	// Phase 3 will retire this once extapi moves out too.
	pkgtenant.Register(func() pkgtenant.TenantRepository {
		return &tenantPkgAdapter{repo: NewTenantRepository(db.Get())}
	})
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

// Module 返回 tenant 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("tenant", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(NewTenantRepository(db.Get())))
		Register(protected, h)
	})
}
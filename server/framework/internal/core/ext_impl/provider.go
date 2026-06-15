package ext_impl

import (
	"context"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/extapi"
)

// defaultProvider is the default extapi.Provider implementation.
//
// Phase 2 status:
//   - user.UserRepository still lives in framework/internal; we read
//     users directly here to avoid an import cycle through extapi.
//   - tenant.TenantRepository has moved to apps/boot/tenant. We don't
//     import apps/ here (internal/ rule); instead apps/boot/tenant
//     registers its factory via the public hook in
//     framework/pkg/tenant (added in Phase 2 alongside pkgauth).
//
// If the tenant module is not loaded, Tenant() facade methods return
// an explicit error so callers know the data layer is missing.
type defaultProvider struct{}

// ----------------- User Facade -----------------
type userFacadeImpl struct {
	repo userRepoRef
}

func (f *userFacadeImpl) GetByID(ctx context.Context, id uint) (*extapi.User, error) {
	u, err := f.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &extapi.User{
		ID: u.ID, TenantID: u.TenantID, AccountID: u.AccountID,
		Code: u.Code, Nickname: u.Nickname, Status: int8(u.Status),
		RealName: u.RealName, Avatar: u.Avatar, Phone: u.Phone,
		Email: u.Email, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}, nil
}

func (f *userFacadeImpl) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]extapi.User, int64, error) {
	users, total, err := f.repo.List(ctx, tenantID, keyword, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]extapi.User, len(users))
	for i, u := range users {
		res[i] = extapi.User{
			ID: u.ID, TenantID: u.TenantID, AccountID: u.AccountID,
			Code: u.Code, Nickname: u.Nickname, Status: int8(u.Status),
			RealName: u.RealName, Avatar: u.Avatar, Phone: u.Phone,
			Email: u.Email, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		}
	}
	return res, total, nil
}

// ----------------- Tenant Facade (Phase 2) -----------------
//
// apps/boot/tenant registers a TenantRepository factory through
// pkg/tenant (the public hook analogous to pkgauth.Register).
// We resolve it lazily per request.
type tenantFacadeImpl struct {
	repoFn func() TenantRepoRef
}

func (f *tenantFacadeImpl) GetByID(ctx context.Context, id uint) (*extapi.Tenant, error) {
	repo := f.repoFn()
	if repo == nil {
		return nil, errTenantNotLoaded
	}
	t, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &extapi.Tenant{
		ID: t.ID, Code: t.Code, Name: t.Name, Status: t.Status,
		Contact: t.Contact, Phone: t.Phone, Email: t.Email,
		Province: t.Province, City: t.City, Area: t.Area, Address: t.Address,
		Config: t.Config, Dashboard: t.Dashboard,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}, nil
}

// ----------------- Provider Methods -----------------
func (p *defaultProvider) User() extapi.UserFacade {
	return &userFacadeImpl{repo: newUserRepoAdapter(db.Get())}
}

func (p *defaultProvider) Tenant() extapi.TenantFacade {
	return &tenantFacadeImpl{repoFn: pkgTenantGet}
}

func InitExtApi() {
	extapi.Set(&defaultProvider{})
}

// errTenantNotLoaded is defined in registry.go
package ext_impl

import (
	"context"
	"time"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/extapi"
	"gx1727.com/xin/framework/pkg/plugin"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

// defaultProvider is the default extapi.Provider implementation.
//
// Phase 3 status:
//   - UserRepository: still served by the historical stop-gap adapter
//     in registry.go (which reads the users table directly). Phase 4
//     migrates user to apps/rbac/user/, and Phase 6 deletes the
//     stop-gap entirely.
//   - TenantRepository: now obtained from the AppContext.Reader via
//     ctx.TenantRepo(). apps/boot/tenant populates it during its
//     Init phase. No more pkgtenant.Get() / pkgtenant.Register().
type defaultProvider struct {
	ctx plugin.Reader
}

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

// ----------------- Tenant Facade (Phase 3) -----------------
//
// apps/boot/tenant publishes its TenantRepository through AppContext
// during Init(). We read it from ctx.TenantRepo() per request.
type tenantFacadeImpl struct {
	repo pkgtenant.TenantRepository
}

func (f *tenantFacadeImpl) GetByID(ctx context.Context, id uint) (*extapi.Tenant, error) {
	if f.repo == nil {
		return nil, errTenantNotLoaded
	}
	t, err := f.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	createdAt, _ := t.CreatedAt.(time.Time)
	updatedAt, _ := t.UpdatedAt.(time.Time)
	return &extapi.Tenant{
		ID: t.ID, Code: t.Code, Name: t.Name, Status: t.Status,
		Contact: t.Contact, Phone: t.Phone, Email: t.Email,
		Province: t.Province, City: t.City, Area: t.Area, Address: t.Address,
		Config: t.Config, Dashboard: t.Dashboard,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

// ----------------- Provider Methods -----------------
func (p *defaultProvider) User() extapi.UserFacade {
	return &userFacadeImpl{repo: newUserRepoAdapter(db.Get())}
}

func (p *defaultProvider) Tenant() extapi.TenantFacade {
	if p.ctx == nil {
		return &tenantFacadeImpl{repo: nil}
	}
	return &tenantFacadeImpl{repo: p.ctx.TenantRepo()}
}

// UpdateTenantCtx keeps the legacy registry.go helpers (used by the
// historical userFacadeImpl path) in sync with the current ctx.
func UpdateTenantCtx(ctx plugin.Reader) { setTenantCtx(ctx) }

// InitExtApi wires the framework's default Provider into the public
// extapi package. Phase 3 takes the AppContext Reader so the tenant
// facade can resolve the repository without a global lookup.
func InitExtApi(ctx plugin.Reader) {
	extapi.Set(&defaultProvider{ctx: ctx})
}

// errTenantNotLoaded is defined in registry.go
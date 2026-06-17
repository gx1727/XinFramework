package ext_impl

import (
	"context"
	"errors"
	"time"

	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/extapi"
	"gx1727.com/xin/framework/pkg/plugin"
)

// defaultProvider is the default extapi.Provider implementation.
//
// Both User and Tenant repositories are obtained from the AppContext.Reader
// via ctx.UserRepo() / ctx.TenantRepo(). apps/rbac/user and apps/boot/tenant
// populate these during their Init phase. No more globals, no more stop-gap
// adapters reading the database directly from here.
type defaultProvider struct {
	ctx plugin.Reader
}

// ----------------- User Facade -----------------
type userFacadeImpl struct {
	repo pkgrbac.UserRepository
}

func (f *userFacadeImpl) GetByID(ctx context.Context, id uint) (*extapi.User, error) {
	if f.repo == nil {
		return nil, errUserNotLoaded
	}
	u, err := f.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toExtAPIUser(u), nil
}

func (f *userFacadeImpl) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]extapi.User, int64, error) {
	if f.repo == nil {
		return nil, 0, errUserNotLoaded
	}
	users, total, err := f.repo.List(ctx, tenantID, keyword, page, size)
	if err != nil {
		return nil, 0, err
	}
	res := make([]extapi.User, len(users))
	for i, u := range users {
		res[i] = *toExtAPIUser(&u)
	}
	return res, total, nil
}

// toExtAPIUser converts the canonical pkgrbac.User (which apps/rbac/user
// aliases as its local User type) to the external API DTO. The two
// structs share the same field set; we coerce optional fields explicitly.
func toExtAPIUser(u *pkgrbac.User) *extapi.User {
	return &extapi.User{
		ID:        u.ID,
		TenantID:  u.TenantID,
		AccountID: u.AccountID,
		Code:      u.Code,
		Nickname:  u.Nickname,
		Status:    u.Status,
		RealName:  u.RealName,
		Avatar:    u.Avatar,
		Phone:     u.Phone,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ----------------- Tenant Facade -----------------
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
	if p.ctx == nil {
		return &userFacadeImpl{repo: nil}
	}
	return &userFacadeImpl{repo: p.ctx.UserRepo()}
}

func (p *defaultProvider) Tenant() extapi.TenantFacade {
	if p.ctx == nil {
		return &tenantFacadeImpl{repo: nil}
	}
	return &tenantFacadeImpl{repo: p.ctx.TenantRepo()}
}

// InitExtApi wires the framework's default Provider into the public
// extapi package. The AppContext Reader carries every cross-module
// repository the Provider needs; nothing is read from globals here.
func InitExtApi(ctx plugin.Reader) {
	extapi.Set(&defaultProvider{ctx: ctx})
}

// errUserNotLoaded and errTenantNotLoaded are returned by the facade
// implementations when the corresponding module is not registered.
var (
	errUserNotLoaded   = errors.New("user module not loaded — register apps/rbac/user in main.go")
	errTenantNotLoaded = errors.New("tenant module not loaded — register apps/boot/tenant in main.go")
)
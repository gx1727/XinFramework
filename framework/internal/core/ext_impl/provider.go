package ext_impl

import (
	"context"

	"gx1727.com/xin/framework/internal/module/tenant"
	"gx1727.com/xin/framework/internal/module/user"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/extapi"
)

type defaultProvider struct{}

// ----------------- User Facade -----------------
type userFacadeImpl struct {
	repo user.UserRepository
}

func (f *userFacadeImpl) GetByID(ctx context.Context, id uint) (*extapi.User, error) {
	u, err := f.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &extapi.User{
		ID: u.ID, TenantID: u.TenantID, AccountID: u.AccountID,
		Code: u.Code, Nickname: u.Nickname, Status: u.Status,
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
			Code: u.Code, Nickname: u.Nickname, Status: u.Status,
			RealName: u.RealName, Avatar: u.Avatar, Phone: u.Phone,
			Email: u.Email, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		}
	}
	return res, total, nil
}

// ----------------- Tenant Facade -----------------
type tenantFacadeImpl struct {
	repo tenant.TenantRepository
}

func (f *tenantFacadeImpl) GetByID(ctx context.Context, id uint) (*extapi.Tenant, error) {
	t, err := f.repo.GetByID(ctx, id)
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
	return &userFacadeImpl{repo: user.NewUserRepository(db.Get())}
}

func (p *defaultProvider) Tenant() extapi.TenantFacade {
	return &tenantFacadeImpl{repo: tenant.NewTenantRepository(db.Get())}
}

func InitExtApi() {
	extapi.Set(&defaultProvider{})
}

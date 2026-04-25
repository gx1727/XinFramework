package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/repository"
	"gx1727.com/xin/framework/pkg/model"
)

type Provider struct {
	db           *pgxpool.Pool
	userRepo     model.UserRepository
	tenantRepo   model.TenantRepository
	accountRepo  model.AccountRepository
	roleRepo     model.RoleRepository
	menuRepo     model.MenuRepository
	resourceRepo model.ResourceRepository
}

var defaultProvider *Provider

func NewProvider(pool *pgxpool.Pool) *Provider {
	return &Provider{
		db:           pool,
		userRepo:     repository.NewUserRepository(pool),
		tenantRepo:   repository.NewTenantRepository(pool),
		accountRepo:  repository.NewAccountRepository(pool),
		roleRepo:     repository.NewRoleRepository(pool),
		menuRepo:     repository.NewMenuRepository(pool),
		resourceRepo: repository.NewResourceRepository(pool),
	}
}

func Init(provider *Provider) {
	defaultProvider = provider
}

func (p *Provider) User() model.UserRepository {
	return p.userRepo
}

func (p *Provider) Tenant() model.TenantRepository {
	return p.tenantRepo
}

func (p *Provider) Account() model.AccountRepository {
	return p.accountRepo
}

func (p *Provider) Role() model.RoleRepository {
	return p.roleRepo
}

func (p *Provider) Menu() model.MenuRepository {
	return p.menuRepo
}

func (p *Provider) Resource() model.ResourceRepository {
	return p.resourceRepo
}

func User() model.UserRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.User()
}

func Tenant() model.TenantRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Tenant()
}

func Account() model.AccountRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Account()
}

func Role() model.RoleRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Role()
}

func Menu() model.MenuRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Menu()
}

func Resource() model.ResourceRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Resource()
}

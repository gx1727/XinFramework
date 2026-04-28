package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/repository"
	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/permission"
)

type Provider struct {
	db              *pgxpool.Pool
	userRepo        model.UserRepository
	tenantRepo      model.TenantRepository
	accountRepo     model.AccountRepository
	accountAuthRepo model.AccountAuthRepository
	roleRepo        model.RoleRepository
	menuRepo        model.MenuRepository
	resourceRepo    model.ResourceRepository
	orgRepo         model.OrganizationRepository
	cmsPostRepo     model.CmsPostRepository
	attachmentRepo  model.AttachmentRepository
	permRepo        *repository.PostgresPermissionRepository
	dsRepo          *repository.PostgresDataScopeRepository
}

var defaultProvider *Provider

func NewProvider(pool *pgxpool.Pool) *Provider {
	return &Provider{
		db:              pool,
		userRepo:        repository.NewUserRepository(pool),
		tenantRepo:      repository.NewTenantRepository(pool),
		accountRepo:     repository.NewAccountRepository(pool),
		accountAuthRepo: repository.NewAccountAuthRepository(pool),
		roleRepo:        repository.NewRoleRepository(pool),
		menuRepo:        repository.NewMenuRepository(pool),
		resourceRepo:    repository.NewResourceRepository(pool),
		orgRepo:         repository.NewOrganizationRepository(pool),
		cmsPostRepo:     repository.NewCmsPostRepository(pool),
		attachmentRepo:  repository.NewAttachmentRepository(pool),
		permRepo:        repository.NewPermissionRepository(pool),
		dsRepo:          repository.NewDataScopeRepository(pool),
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

func (p *Provider) AccountAuth() model.AccountAuthRepository {
	return p.accountAuthRepo
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

func (p *Provider) Organization() model.OrganizationRepository {
	return p.orgRepo
}

func (p *Provider) CmsPost() model.CmsPostRepository {
	return p.cmsPostRepo
}

func (p *Provider) Attachment() model.AttachmentRepository {
	return p.attachmentRepo
}

func (p *Provider) Permission() permission.UserPermissionRepository {
	return p.permRepo
}

func (p *Provider) DataScope() permission.DataScopeRepository {
	return p.dsRepo
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

func AccountAuth() model.AccountAuthRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.AccountAuth()
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

func Organization() model.OrganizationRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Organization()
}

func CmsPost() model.CmsPostRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.CmsPost()
}

func Attachment() model.AttachmentRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Attachment()
}

func Permission() permission.UserPermissionRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.Permission()
}

func DataScope() permission.DataScopeRepository {
	if defaultProvider == nil {
		return nil
	}
	return defaultProvider.DataScope()
}

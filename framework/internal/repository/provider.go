package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/permission"
	pkg_repository "gx1727.com/xin/framework/pkg/repository"
)

// 确保 Provider 实现 pkg/repository.Provider
var _ pkg_repository.Provider = (*Provider)(nil)

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
	permRepo        *PostgresPermissionRepository
	dsRepo          *PostgresDataScopeRepository
}

var defaultProvider *Provider

func NewProvider(pool *pgxpool.Pool) *Provider {
	return &Provider{
		db:              pool,
		userRepo:        NewUserRepository(pool),
		tenantRepo:      NewTenantRepository(pool),
		accountRepo:     NewAccountRepository(pool),
		accountAuthRepo: NewAccountAuthRepository(pool),
		roleRepo:        NewRoleRepository(pool),
		menuRepo:        NewMenuRepository(pool),
		resourceRepo:    NewResourceRepository(pool),
		orgRepo:         NewOrganizationRepository(pool),
		cmsPostRepo:     NewCmsPostRepository(pool),
		attachmentRepo:  NewAttachmentRepository(pool),
		permRepo:        NewPermissionRepository(pool),
		dsRepo:          NewDataScopeRepository(pool),
	}
}

func Init(provider *Provider) {
	defaultProvider = provider
	pkg_repository.SetProvider(provider)
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

// GetDefaultProvider 获取全局默认的 Provider 实例（供外部插件使用）
func GetDefaultProvider() *Provider {
	return defaultProvider
}

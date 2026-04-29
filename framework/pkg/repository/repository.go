package repository

import (
	"gx1727.com/xin/framework/pkg/model"
	"gx1727.com/xin/framework/pkg/permission"
)

// Provider 定义了聚合所有核心数据访问能力的接口，供外部插件调用
type Provider interface {
	User() model.UserRepository
	Tenant() model.TenantRepository
	Account() model.AccountRepository
	AccountAuth() model.AccountAuthRepository
	Role() model.RoleRepository
	Menu() model.MenuRepository
	Resource() model.ResourceRepository
	Organization() model.OrganizationRepository
	CmsPost() model.CmsPostRepository
	Attachment() model.AttachmentRepository
	Permission() permission.UserPermissionRepository
	DataScope() permission.DataScopeRepository
}

var globalProvider Provider

// SetProvider 设置全局 Provider 实例
func SetProvider(p Provider) {
	globalProvider = p
}

// Get 获取全局 Provider 实例
func Get() Provider {
	return globalProvider
}

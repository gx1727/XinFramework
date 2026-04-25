package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/repository"
	"gx1727.com/xin/framework/pkg/model"
)

// 全局 Repository 实例
var (
	_userRepo     model.UserRepository
	_tenantRepo   model.TenantRepository
	_accountRepo  model.AccountRepository
	_roleRepo     model.RoleRepository
	_menuRepo     model.MenuRepository
	_resourceRepo model.ResourceRepository
	_dbPool       *pgxpool.Pool
)

// Init 初始化 Repository（在框架启动时调用）
func Init(pool *pgxpool.Pool) {
	_dbPool = pool
	_userRepo = repository.NewUserRepository(pool)
	_tenantRepo = repository.NewTenantRepository(pool)
	_accountRepo = repository.NewAccountRepository(pool)
	_roleRepo = repository.NewRoleRepository(pool)
	_menuRepo = repository.NewMenuRepository(pool)
	_resourceRepo = repository.NewResourceRepository(pool)
}

// User 返回 UserRepository
func User() model.UserRepository {
	return _userRepo
}

// Tenant 返回 TenantRepository
func Tenant() model.TenantRepository {
	return _tenantRepo
}

// Account 返回 AccountRepository
func Account() model.AccountRepository {
	return _accountRepo
}

// Role 返回 RoleRepository
func Role() model.RoleRepository {
	return _roleRepo
}

// Menu 返回 MenuRepository
func Menu() model.MenuRepository {
	return _menuRepo
}

// Resource 返回 ResourceRepository
func Resource() model.ResourceRepository {
	return _resourceRepo
}

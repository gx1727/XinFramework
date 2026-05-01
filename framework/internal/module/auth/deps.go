package auth

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/module/role"
	"gx1727.com/xin/framework/internal/module/tenant"
	"gx1727.com/xin/framework/internal/module/user"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/session"
)

type SessionManager interface {
	Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
	Revoke(sessionID string) error
}

type Dependencies struct {
	DB          *pgxpool.Pool
	Config      *config.Config
	Session     SessionManager
	AccountRepo AccountRepository
	TenantRepo  tenant.TenantRepository
	RoleRepo    role.RoleRepository
	UserRepo    user.UserRepository
}

type defaultSessionManager struct{}

func (defaultSessionManager) Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	return session.Create(sessionID, userID, tenantID, role, ttl)
}

func (defaultSessionManager) Revoke(sessionID string) error {
	return session.Revoke(sessionID)
}

func DefaultDependencies(cfg *config.Config, db *pgxpool.Pool, repos Repositories) Dependencies {
	return Dependencies{
		DB:          db,
		Config:      cfg,
		Session:     defaultSessionManager{},
		AccountRepo: repos.Account,
		TenantRepo:  repos.Tenant,
		RoleRepo:    repos.Role,
		UserRepo:    repos.User,
	}
}

// Repositories 鉴权模块所需的仓储接口
type Repositories struct {
	Account AccountRepository
	Tenant  tenant.TenantRepository
	Role    role.RoleRepository
	User    user.UserRepository
}

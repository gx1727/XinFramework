package auth

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/model"
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
	AccountRepo model.AccountRepository
	TenantRepo  model.TenantRepository
	RoleRepo    model.RoleRepository
	UserRepo    model.UserRepository
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

type Repositories struct {
	Account model.AccountRepository
	Tenant  model.TenantRepository
	Role    model.RoleRepository
	User    model.UserRepository
}

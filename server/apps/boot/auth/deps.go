package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/login_security"
	"gx1727.com/xin/framework/pkg/session"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
)

type SessionManager interface {
	Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
	Revoke(sessionID string) error
}

// PlatformRoleRepository 平台级角色访问接口（最小子集，登录阶段使用）
type PlatformRoleRepository interface {
	GetRolesByUserID(ctx context.Context, userID uint) ([]string, error)
	// GetRolesByAccountID 直接按 account_id 查（用于 platform-login，
	// 此时 user 可能未绑 user 行）
	GetRolesByAccountID(ctx context.Context, accountID uint) ([]string, error)
}

type Dependencies struct {
	DB            *pgxpool.Pool
	Config        *config.Config
	Session       SessionManager
	AccountRepo   AccountRepository
	TenantRepo    pkgtenant.TenantRepository
	PlatformRepo  PlatformRoleRepository
	Security      *login_security.SecurityService // 可为 nil（未装配 login_security 时降级为 noop）
}

type defaultSessionManager struct{}

func (defaultSessionManager) Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error {
	return session.Create(sessionID, userID, tenantID, role, ttl)
}

func (defaultSessionManager) Revoke(sessionID string) error {
	return session.Revoke(sessionID)
}

func DefaultDependencies(cfg *config.Config, db *pgxpool.Pool, repos Repositories, security *login_security.SecurityService) Dependencies {
	return Dependencies{
		DB:           db,
		Config:       cfg,
		Session:      defaultSessionManager{},
		AccountRepo:  repos.Account,
		TenantRepo:   repos.Tenant,
		PlatformRepo: repos.Platform,
		Security:     security,
	}
}

type Repositories struct {
	Account  AccountRepository
	Tenant   pkgtenant.TenantRepository
	Platform PlatformRoleRepository
}

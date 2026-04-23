package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
)

type Service struct {
	db      *pgxpool.Pool
	config  *config.Config
	session SessionManager
}

func NewService(deps Dependencies) *Service {
	return &Service{
		db:      deps.DB,
		config:  deps.Config,
		session: deps.Session,
	}
}

func (s *Service) Login(ctx context.Context, req loginRequest) (*loginResult, error) {
	identity, err := ResolveLoginIdentity(ctx, s.db, req.Account, req.TenantID)
	if err != nil {
		switch {
		case errors.Is(err, ErrBackendUnavailable):
			return nil, ErrBackendUnavailable
		case errors.Is(err, ErrAccountNotFound):
			return nil, ErrInvalidAccountOrPassword
		case errors.Is(err, ErrTenantBindingNotFound):
			return nil, ErrTenantBindingNotFound
		default:
			return nil, ErrInvalidAccountOrPassword
		}
	}

	ok, err := verifyPassword(identity.PasswordHash, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidAccountOrPassword
	}
	if identity.UserStatus != 1 {
		return nil, ErrUserDisabled
	}

	if s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	sessionID := uuid.NewString()
	if err := s.session.Create(sessionID, identity.UserID, identity.TenantID, identity.RoleCode, time.Duration(s.config.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&s.config.JWT, identity.UserID, identity.TenantID, identity.RoleCode, sessionID)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	res := &loginResult{Token: token}
	res.User.ID = identity.UserID
	res.User.TenantID = identity.TenantID
	res.User.Code = identity.UserCode
	res.User.Role = identity.RoleCode
	return res, nil
}

func (s *Service) Logout(sessionID string) error {
	if s.config == nil || s.session == nil {
		return ErrBackendUnavailable
	}
	if sessionID == "" {
		return ErrInvalidToken
	}
	if err := s.session.Revoke(sessionID); err != nil {
		return ErrSessionRevokeFailed
	}
	return nil
}

func (s *Service) Register(ctx context.Context, req registerRequest) (*registerResult, error) {
	if s.db == nil || s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM accounts 
			WHERE is_deleted = FALSE 
			AND (phone = $1 OR email = $1)
		)`, req.Account).Scan(&exists)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	if exists {
		return nil, ErrAccountAlreadyExists
	}

	var tenantStatus int16
	err = s.db.QueryRow(ctx, `
		SELECT status FROM tenants 
		WHERE is_deleted = FALSE AND id = $1`, req.TenantID).Scan(&tenantStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, ErrRegisterFailed
	}
	if tenantStatus != 1 {
		return nil, ErrTenantNotFound
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, ErrRegisterFailed
	}

	var newAccountID uint
	var newUserID uint
	var newUserCode string

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO accounts (phone, email, username, password, real_name) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`, req.Account, req.Account, req.Account, passwordHash, req.RealName).Scan(&newAccountID)
	if err != nil {
		return nil, ErrRegisterFailed
	}

	newUserCode = uuid.NewString()[:8]
	err = tx.QueryRow(ctx, `
		INSERT INTO users (tenant_id, account_id, code, status) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`, req.TenantID, newAccountID, newUserCode, 1).Scan(&newUserID)
	if err != nil {
		return nil, ErrRegisterFailed
	}

	var roleID uint
	err = tx.QueryRow(ctx, `
		SELECT id FROM roles 
		WHERE is_deleted = FALSE AND tenant_id = $1 AND is_default = TRUE 
		LIMIT 1`, req.TenantID).Scan(&roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDefaultRoleNotFound
		}
		return nil, ErrRegisterFailed
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (tenant_id, user_id, role_id) 
		VALUES ($1, $2, $3)`, req.TenantID, newUserID, roleID)
	if err != nil {
		return nil, ErrRegisterFailed
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, ErrRegisterFailed
	}

	sessionID := uuid.NewString()
	if err := s.session.Create(sessionID, newUserID, req.TenantID, "user", time.Duration(s.config.JWT.Expire)*time.Second); err != nil {
		return nil, ErrSessionCreateFailed
	}

	token, err := jwtpkg.Generate(&s.config.JWT, newUserID, req.TenantID, "user", sessionID)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	res := &registerResult{Token: token}
	res.User.ID = newUserID
	res.User.TenantID = req.TenantID
	res.User.Code = newUserCode
	res.User.Role = "user"
	return res, nil
}

package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

type tokenPair struct {
	accessToken  string
	refreshToken string
}

func (s *Service) generateTokens(ctx context.Context, userID, tenantID uint, role string) (*tokenPair, error) {
	if s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	sessionID := uuid.NewString()
	refreshTTL := time.Duration(s.config.JWT.RefreshExpire) * time.Second

	if err := s.session.Create(sessionID, userID, tenantID, role, refreshTTL); err != nil {
		return nil, ErrSessionCreateFailed
	}

	accessToken, err := jwtpkg.Generate(&s.config.JWT, userID, tenantID, role, sessionID)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	refreshToken, err := jwtpkg.GenerateWithType(&s.config.JWT, userID, tenantID, role, sessionID, jwtpkg.TokenTypeRefresh)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	return &tokenPair{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

func (s *Service) Login(ctx context.Context, req loginRequest) (*LoginResult, error) {
	identity, err := ResolveLoginIdentity(ctx, s.db, req.Account, req.TenantID)
	if err != nil {
		switch {
		case errors.Is(err, ErrBackendUnavailable):
			return nil, ErrBackendUnavailable
		case errors.Is(err, errAccountNotFound):
			return nil, ErrInvalidAccountOrPassword
		case errors.Is(err, errTenantBindingNotFound):
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

	tokens, err := s.generateTokens(ctx, identity.UserID, identity.TenantID, identity.RoleCode)
	if err != nil {
		return nil, err
	}

	res := &LoginResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
	}
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

func (s *Service) Refresh(ctx context.Context, req refreshRequest) (*refreshResult, error) {
	if s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	claims, err := jwtpkg.ValidateRefresh(req.RefreshToken, &s.config.JWT)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	newTokens, err := s.generateTokens(ctx, claims.UserID, claims.TenantID, claims.Role)
	if err != nil {
		return nil, err
	}

	if err := s.session.Revoke(claims.SessionID); err != nil {
		// Ignore revoke error, new session is already created
	}

	return &refreshResult{
		Token:        newTokens.accessToken,
		RefreshToken: newTokens.refreshToken,
	}, nil
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

	_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(req.TenantID)))
	if err != nil {
		return nil, fmt.Errorf("set tenant_id: %w", err)
	}

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

	tokens, err := s.generateTokens(ctx, newUserID, req.TenantID, "user")
	if err != nil {
		return nil, err
	}

	res := &registerResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
	}
	res.User.ID = newUserID
	res.User.TenantID = req.TenantID
	res.User.Code = newUserCode
	res.User.Role = "user"
	return res, nil
}

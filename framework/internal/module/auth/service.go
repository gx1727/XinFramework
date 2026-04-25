package auth

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
	"gx1727.com/xin/framework/pkg/model"
)

type LoginIdentity struct {
	UserID       uint
	TenantID     uint
	UserCode     string
	UserStatus   int16
	RoleCode     string
	PasswordHash string
}

func ResolveLoginIdentity(ctx context.Context, d *pgxpool.Pool, account string, tenantID uint) (*LoginIdentity, error) {
	if d == nil {
		return nil, ErrBackendUnavailable
	}

	tx, err := d.Begin(ctx)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	defer tx.Rollback(ctx)

	if tenantID > 0 {
		_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID)))
		if err != nil {
			return nil, fmt.Errorf("set tenant_id: %w", err)
		}
	} else {
		_, err = tx.Exec(ctx, "SELECT set_config('app.mode', $1, true)", "single")
		if err != nil {
			return nil, fmt.Errorf("set mode: %w", err)
		}
	}

	var accID uint
	var password string
	err = tx.QueryRow(ctx, `
		SELECT id, password
		FROM accounts
		WHERE username = $1 OR phone = $1 OR email = $1
		LIMIT 1`, account).Scan(&accID, &password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errAccountNotFound
		}
		return nil, err
	}

	query := `
		SELECT id, tenant_id, code, status
		FROM users
		WHERE  account_id = $1`
	args := []interface{}{accID}

	query += " ORDER BY id ASC LIMIT 1"

	var uID uint
	var uTenantID uint
	var uCode string
	var uStatus int16
	err = tx.QueryRow(ctx, query, args...).Scan(&uID, &uTenantID, &uCode, &uStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errTenantBindingNotFound
		}
		return nil, err
	}

	if tenantID == 0 {
		_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(uTenantID)))
		if err != nil {
			return nil, fmt.Errorf("set tenant_id: %w", err)
		}

		_, err = tx.Exec(ctx, "SELECT set_config('app.mode', $1, true)", "saas")
		if err != nil {
			return nil, fmt.Errorf("set mode: %w", err)
		}
	}

	roleCode := "user"
	err = tx.QueryRow(ctx, `
		SELECT r.code
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY ur.id ASC LIMIT 1`, uID).Scan(&roleCode)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, ErrRegisterFailed
	}

	return &LoginIdentity{
		UserID:       uID,
		TenantID:     uTenantID,
		UserCode:     uCode,
		UserStatus:   uStatus,
		RoleCode:     roleCode,
		PasswordHash: password,
	}, nil
}

type Service struct {
	db          *pgxpool.Pool
	config      *config.Config
	session     SessionManager
	accountRepo model.AccountRepository
	tenantRepo  model.TenantRepository
	roleRepo    model.RoleRepository
	userRepo    model.UserRepository
}

func NewService(deps Dependencies) *Service {
	return &Service{
		db:          deps.DB,
		config:      deps.Config,
		session:     deps.Session,
		accountRepo: deps.AccountRepo,
		tenantRepo:  deps.TenantRepo,
		roleRepo:    deps.RoleRepo,
		userRepo:    deps.UserRepo,
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

	exists, err := s.accountRepo.Exists(ctx, req.Account)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	if exists {
		return nil, ErrAccountAlreadyExists
	}

	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		if errors.Is(err, model.ErrTenantNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, ErrRegisterFailed
	}
	if tenant.Status != 1 {
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

	newAccount, err := s.accountRepo.Create(ctx, req.Account, req.Account, req.Account, req.RealName, passwordHash)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	newAccountID = newAccount.ID

	newUserCode, err = generateUserCode(ctx, s.db, req.TenantID, UserCodeFormatSequential)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO users (tenant_id, account_id, code, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, req.TenantID, newAccountID, newUserCode, 1).Scan(&newUserID)
	if err != nil {
		return nil, ErrRegisterFailed
	}

	roles, err := s.roleRepo.List(ctx, req.TenantID)
	if err != nil {
		return nil, ErrRegisterFailed
	}
	var roleID uint
	for _, role := range roles {
		if role.IsDefault {
			roleID = role.ID
			break
		}
	}
	if roleID == 0 {
		return nil, ErrDefaultRoleNotFound
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

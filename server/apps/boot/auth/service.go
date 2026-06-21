package auth

import (
	"gx1727.com/xin/apps/admin/platform_tenant"

	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
)

type LoginIdentity struct {
	UserID       uint
	TenantID     uint
	UserCode     string
	UserStatus   int16
	RoleCode     string
	PasswordHash string

	// 用户展示资料（侧边栏 / NavUser 用），来自 users JOIN accounts
	Nickname string
	RealName string
	Avatar   string
	Email    string
}

func ResolveLoginIdentity(ctx context.Context, d *pgxpool.Pool, account string, tenantID uint) (*LoginIdentity, error) {
	if d == nil {
		return nil, ErrBackendUnavailable
	}
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	var identity LoginIdentity

	err := db.RunInTenantTx(ctx, d, tenantID, func(ctx context.Context) error {
		querier, err := db.GetQuerier(ctx, d)
		if err != nil {
			return err
		}

		var accID uint
		var password string
		err = querier.QueryRow(ctx, `
			SELECT id, password
			FROM accounts
			WHERE username = $1 OR phone = $1 OR email = $1
			LIMIT 1`, account).Scan(&accID, &password)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errAccountNotFound
			}
			return err
		}

		var (
			uID       uint
			uTenantID uint
			uCode     string
			uStatus   int16
			uNickname string
			uRealName string
			uAvatar   string
			aEmail    string
		)
		err = querier.QueryRow(ctx, `
			SELECT u.id, u.tenant_id, u.code, u.status,
			       COALESCE(u.nickname, ''), COALESCE(u.real_name, ''), COALESCE(u.avatar, ''),
			       COALESCE(a.email, '')
			FROM users u
			JOIN accounts a ON a.id = u.account_id
			WHERE u.account_id = $1
			ORDER BY u.id ASC LIMIT 1`, accID).Scan(
			&uID, &uTenantID, &uCode, &uStatus,
			&uNickname, &uRealName, &uAvatar, &aEmail,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errTenantBindingNotFound
			}
			return err
		}

		roleCode := "user"
		err = querier.QueryRow(ctx, `
			SELECT r.code
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1
			ORDER BY ur.id ASC LIMIT 1`, uID).Scan(&roleCode)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			// ignore role not found
		}

		identity = LoginIdentity{
			UserID:       uID,
			TenantID:     uTenantID,
			UserCode:     uCode,
			UserStatus:   uStatus,
			RoleCode:     roleCode,
			PasswordHash: password,
			Nickname:     uNickname,
			RealName:     uRealName,
			Avatar:       uAvatar,
			Email:        aEmail,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &identity, nil
}

// hasPlatformRole 判断 userID 是否拥有指定的平台级角色（如 super_admin）。
// account_roles 不受 RLS 限制，可直接在事务外查。
func (s *Service) hasPlatformRole(ctx context.Context, userID uint, role string) bool {
	if s.platformRp == nil || userID == 0 {
		return false
	}
	roles, err := s.platformRp.GetRolesByUserID(ctx, userID)
	if err != nil {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

type Service struct {
	db          *pgxpool.Pool
	config      *config.Config
	session     SessionManager
	accountRepo AccountRepository
	tenantRepo  platformtenant.TenantRepository
	platformRp  PlatformRoleRepository
}

func NewService(deps Dependencies) *Service {
	return &Service{
		db:          deps.DB,
		config:      deps.Config,
		session:     deps.Session,
		accountRepo: deps.AccountRepo,
		tenantRepo:  deps.TenantRepo,
		platformRp:  deps.PlatformRepo,
	}
}

type tokenPair struct {
	accessToken   string
	refreshToken  string
	platformRoles []string
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

	// 取出用户绑定的平台级角色（如 super_admin），写入 JWT
	platformRoles := s.loadPlatformRoles(ctx, userID)

	accessToken, err := jwtpkg.GenerateWithPlatformRoles(&s.config.JWT, userID, tenantID, role, sessionID, platformRoles, jwtpkg.TokenTypeAccess)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	refreshToken, err := jwtpkg.GenerateWithPlatformRoles(&s.config.JWT, userID, tenantID, role, sessionID, platformRoles, jwtpkg.TokenTypeRefresh)
	if err != nil {
		return nil, ErrGenerateTokenFailed
	}

	return &tokenPair{
		accessToken:   accessToken,
		refreshToken:  refreshToken,
		platformRoles: platformRoles,
	}, nil
}

// loadPlatformRoles 查 user_id 对应的 account_id 拥有的平台角色
// 在普通租户事务外查（account_roles 不受 RLS 限制）。
func (s *Service) loadPlatformRoles(ctx context.Context, userID uint) []string {
	if s.platformRp == nil || s.db == nil || userID == 0 {
		return nil
	}
	roles, err := s.platformRp.GetRolesByUserID(ctx, userID)
	if err != nil {
		return nil
	}
	return roles
}

// Login 租户域登录（业务用户）。需要传 tenant_id；user 必须绑到该 tenant。
func (s *Service) Login(ctx context.Context, req tenantLoginRequest) (*LoginResult, error) {
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
		Scope:        LoginScopeTenant,
	}
	res.User.ID = identity.UserID
	res.User.TenantID = identity.TenantID
	res.User.Code = identity.UserCode
	res.User.Role = identity.RoleCode
	res.User.Nickname = identity.Nickname
	res.User.RealName = identity.RealName
	res.User.Avatar = identity.Avatar
	res.User.Email = identity.Email
	res.User.PlatformRoles = tokens.platformRoles
	return res, nil
}

// PlatformLogin 平台域登录（super_admin 等）。
//
// 不传 tenant_id：平台管理员不属于任何租户。
// 验证流程：
//  1. accounts 表（不受 RLS）查账号 + 验证密码
//  2. 账号状态必须启用
//  3. 至少有一个 platform_role（如 super_admin）
//  4. 生成 token，tenant_id 固定 0
//
// 成功后 token 仅能访问 /api/v1/platform/* 平台域路由。
func (s *Service) PlatformLogin(ctx context.Context, req platformLoginRequest) (*LoginResult, error) {
	if s.db == nil || s.accountRepo == nil || s.platformRp == nil {
		return nil, ErrBackendUnavailable
	}

	// 1. accounts 表查账号
	passwordHash, accountID, accountStatus, err := s.accountRepo.GetPasswordAndStatus(ctx, req.Account)
	if err != nil {
		if errors.Is(err, errAccountNotFound) {
			return nil, ErrPlatformLoginAccountNotFound
		}
		return nil, ErrBackendUnavailable
	}
	if accountStatus != 1 {
		return nil, ErrPlatformLoginDisabled
	}

	// 2. 验密码
	ok, err := verifyPassword(passwordHash, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidAccountOrPassword
	}

	// 3. 查 platform_roles
	platformRoles, err := s.platformRp.GetRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, ErrBackendUnavailable
	}
	if len(platformRoles) == 0 {
		return nil, ErrPlatformLoginNotAdmin
	}

	// 4. 生成 token，tenant_id = 0（platform 域语义）
	//    role 用 "_platform" 占位（不在 tenants / roles 表里）
	tokens, err := s.generateTokens(ctx, 0, 0, "_platform")
	if err != nil {
		return nil, err
	}

	res := &LoginResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
		Scope:        LoginScopePlatform,
		User: User{
			ID:            accountID, // 这里 ID 是 account_id（不是 user_id），因为平台用户可能没绑 user
			TenantID:      0,
			Code:          req.Account,
			Role:          "_platform",
			PlatformRoles: platformRoles,
		},
	}
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

	var newUserID uint
	var newUserCode string

	err := db.RunInTenantTx(ctx, s.db, req.TenantID, func(ctx context.Context) error {
		exists, err := s.accountRepo.Exists(ctx, req.Account)
		if err != nil {
			return ErrRegisterFailed
		}
		if exists {
			return ErrAccountAlreadyExists
		}

		t, err := s.tenantRepo.GetByID(ctx, req.TenantID)
		if err != nil {
			if errors.Is(err, platformtenant.ErrTenantNotFoundDB) {
				return ErrTenantNotFound
			}
			return ErrRegisterFailed
		}
		if t.Status != 1 {
			return ErrTenantNotFound
		}

		passwordHash, err := HashPassword(req.Password)
		if err != nil {
			return ErrRegisterFailed
		}

		newAccount, err := s.accountRepo.Create(ctx, req.Account, req.Account, req.Account, req.RealName, passwordHash)
		if err != nil {
			return ErrRegisterFailed
		}

		querier, err := db.GetQuerier(ctx, s.db)
		if err != nil {
			return ErrRegisterFailed
		}
		tx, ok := querier.(pgx.Tx)
		if !ok {
			return ErrRegisterFailed
		}

		newUserCode, err = generateUserCode(ctx, tx, req.TenantID, UserCodeFormatSequential)
		if err != nil {
			return ErrRegisterFailed
		}

		err = querier.QueryRow(ctx, `
			INSERT INTO users (tenant_id, account_id, code, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id`, req.TenantID, newAccount.ID, newUserCode, 1).Scan(&newUserID)
		if err != nil {
			return ErrRegisterFailed
		}

		var roleID uint
		err = querier.QueryRow(ctx, `
			SELECT id FROM roles
			WHERE is_deleted = FALSE AND tenant_id = $1 AND is_default = TRUE
			LIMIT 1
		`, req.TenantID).Scan(&roleID)
		if err != nil {
			return ErrDefaultRoleNotFound
		}

		_, err = querier.Exec(ctx, `
			INSERT INTO user_roles (tenant_id, user_id, role_id)
			VALUES ($1, $2, $3)`, req.TenantID, newUserID, roleID)
		if err != nil {
			return ErrRegisterFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(ctx, newUserID, req.TenantID, "user")
	if err != nil {
		return nil, err
	}

	res := &registerResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
		Scope:        LoginScopeTenant,
	}
	res.User.ID = newUserID
	res.User.TenantID = req.TenantID
	res.User.Code = newUserCode
	res.User.Role = "user"
	res.User.RealName = req.RealName
	res.User.PlatformRoles = tokens.platformRoles
	// nickname/avatar/email 暂未在注册时收集，留空字符串（DB 列也未填）
	return res, nil
}

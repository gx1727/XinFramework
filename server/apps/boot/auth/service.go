package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgauth "gx1727.com/xin/framework/pkg/auth"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/logger"
	"gx1727.com/xin/framework/pkg/login_security"
	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/xincontext"
)

type LoginIdentity struct {
	AccountID    xincontext.AccountID // accounts.id（用于 login_security.history）
	UserID       xincontext.UserID    // tenant_users.id
	TenantID     xincontext.TenantID
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

func ResolveLoginIdentity(ctx context.Context, d *pgxpool.Pool, account string, tenantID xincontext.TenantID) (*LoginIdentity, error) {
	if d == nil {
		return nil, ErrBackendUnavailable
	}
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	var identity LoginIdentity

	err := db.RunInTenantTx(ctx, d, uint(tenantID), func(ctx context.Context) error {
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
			uID       xincontext.UserID
			uTenantID xincontext.TenantID
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
			FROM tenant_users u
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

		roleCode := RoleCodeUser
		// 用户没有任何角色时 SELECT 返回 NoRows,留 roleCode 兜底值 "user" 继续走流程;
		// 其它 DB 错误必须返回,不能像之前那样被空 if 块吞掉。
		err = querier.QueryRow(ctx, `
			SELECT r.code
			FROM tenant_user_roles ur
			JOIN tenant_roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1
			ORDER BY ur.id ASC LIMIT 1`, uID).Scan(&roleCode)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("load user role: %w", err)
		}

		identity = LoginIdentity{
			AccountID:    xincontext.NewAccountID(accID),
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
	tenantRepo  pkgtenant.TenantRepository
	platformRp  PlatformRoleRepository
	security    *login_security.SecurityService // 可为 nil；为 nil 时跳过锁定 / 异地检测
}

func NewService(deps Dependencies) *Service {
	return &Service{
		db:          deps.DB,
		config:      deps.Config,
		session:     deps.Session,
		accountRepo: deps.AccountRepo,
		tenantRepo:  deps.TenantRepo,
		platformRp:  deps.PlatformRepo,
		security:    deps.Security,
	}
}

// attemptFromContext 从 ctx 中提取登录尝试所需的请求元数据（IP/UA/DeviceID）。
//
// Auth 中间件已经把 IP/UA/DeviceID 注入到 XinContext 里（xincontext.ContextFrom）。
// 登录流程的特殊性：登录时 ctx 里通常没有 UserContext（没登录），所以直接从 XinContext 读。
func attemptFromContext(ctx context.Context) (ip, ua, deviceID string) {
	xc, ok := xincontext.XinContextFrom(ctx)
	if !ok || xc == nil {
		return "", "", ""
	}
	return xc.IP, xc.UserAgent, xc.DeviceID
}

// checkAccountLock 在登录前检查账号是否被锁。返回 nil 表示可继续登录；非 nil 表示被锁。
func (s *Service) checkAccountLock(ctx context.Context, account string) error {
	if s.security == nil {
		return nil
	}
	lock, err := s.security.CheckLock(ctx, account)
	if err != nil {
		// 后端错误不阻塞登录——fallback 到"未锁"让用户能进
		logger.Module("auth").Warnf("CheckLock failed for account=%s: %v", account, err)
		return nil
	}
	if lock == nil {
		return nil
	}
	return ErrAccountLocked
}

// recordFailure 在登录失败后调用；触发锁定（必要时）+ 记录 attempt。
func (s *Service) recordFailure(ctx context.Context, account string, scope login_security.Scope, tenantID uint, reason login_security.FailureReason) {
	if s.security == nil {
		return
	}
	ip, ua, _ := attemptFromContext(ctx)
	count, triggered, err := s.security.RecordFailure(ctx, account, ip, ua, reason, scope, tenantID)
	if err != nil {
		logger.Module("auth").Warnf("RecordFailure failed account=%s: %v", account, err)
		return
	}
	logger.Module("auth").Debugf("login failed: account=%s scope=%s count=%d/%d triggeredLock=%v",
		account, scope, count, s.securityCfg().MaxFailedAttempts, triggered)
}

// recordSuccess 在登录成功后调用：写 login_history + 触发异地告警（必要时）。
//
// sessionID 关联到 auth_sessions.id（便于后续通过 session 反查登录来源）。
// deviceID 来自 ctx 的 X-Device-ID header（前端可选择性设置）。
func (s *Service) recordSuccess(
	ctx context.Context,
	accountID uint,
	userID uint,
	tenantID uint,
	scope login_security.Scope,
	sessionID string,
	role string,
	platformRoles []string,
) {
	if s.security == nil {
		return
	}
	ip, ua, deviceID := attemptFromContext(ctx)
	entry := login_security.LoginHistoryEntry{
		AccountID: accountID,
		UserID:    userID,
		TenantID:  tenantID,
		Scope:     scope,
		IP:        ip,
		UserAgent: ua,
		DeviceID:  deviceID,
		SessionID: sessionID,
		LoginAt:   time.Now(),
	}
	if _, _, err := s.security.RecordSuccess(ctx, entry); err != nil {
		// history 写入失败不应阻断登录——已经登录成功了
		logger.Module("auth").Warnf("RecordSuccess history failed: accountID=%d ip=%s err=%v", accountID, ip, err)
	}
}

// securityCfg 把 config 里的分钟数转换为 SecurityConfig（懒计算）。
func (s *Service) securityCfg() login_security.SecurityConfig {
	if s.config == nil {
		return login_security.DefaultSecurityConfig()
	}
	ls := s.config.LoginSecurity
	if !ls.Enabled {
		return login_security.SecurityConfig{Enabled: false}
	}
	c := login_security.DefaultSecurityConfig()
	c.Enabled = true
	if ls.MaxFailedAttempts > 0 {
		c.MaxFailedAttempts = ls.MaxFailedAttempts
	}
	if ls.LockDurationMin > 0 {
		c.LockDuration = time.Duration(ls.LockDurationMin) * time.Minute
	}
	if ls.FailureWindowMin > 0 {
		c.FailureWindow = time.Duration(ls.FailureWindowMin) * time.Minute
	}
	if ls.IPFailureThreshold > 0 {
		c.IPFailureThreshold = ls.IPFailureThreshold
	}
	if ls.IPFailureWindowMin > 0 {
		c.IPFailureWindow = time.Duration(ls.IPFailureWindowMin) * time.Minute
	}
	if ls.AnomalyHistoryLimit > 0 {
		c.AnomalyHistoryLimit = ls.AnomalyHistoryLimit
	}
	c.AnomalyDeviceMatch = ls.AnomalyDeviceMatch
	c.AnomalyNotifyInSite = ls.AnomalyNotifyInSite
	c.AnomalyNotifyEmail = ls.AnomalyNotifyEmail
	c.AnomalyNotifySMS = ls.AnomalyNotifySMS
	c.LockNotifyInSite = ls.LockNotifyInSite
	c.LockNotifyEmail = ls.LockNotifyEmail
	c.LockNotifySMS = ls.LockNotifySMS
	return c
}

type tokenPair struct {
	accessToken   string
	refreshToken  string
	sessionID     string // session_id（写入 login_history 用）
	platformRoles []string
}

// generateTokens 签发 access + refresh JWT，写入 session。
//
// userID 是 JWT 的 Subject：tenant user = users.id；platform admin = account_id（与 LoginResult.User.ID 一致）。
// accountID 用于查 account_roles：tenant user 走 userID 路径（users.account_id 中转），platform admin 走 accountID 直查。
// 两个都给比"二选一"更鲁棒，避免 PlatformLogin 漏签 platformRoles 导致后续 RequirePlatformRole 失败。
func (s *Service) generateTokens(ctx context.Context, userID, tenantID uint, role string, accountID uint) (*tokenPair, error) {
	if s.config == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	sessionID := uuid.NewString()
	refreshTTL := time.Duration(s.config.JWT.RefreshExpire) * time.Second

	if err := s.session.Create(sessionID, userID, tenantID, role, refreshTTL); err != nil {
		return nil, ErrSessionCreateFailed
	}

	// 取出用户绑定的平台级角色（如 super_admin），写入 JWT
	platformRoles := s.loadPlatformRoles(ctx, userID, accountID)

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
		sessionID:     sessionID,
		platformRoles: platformRoles,
	}, nil
}

// loadPlatformRoles 查 account_id 拥有的平台角色。
//
// 优先走 user_id 路径（user_id → users.account_id → account_roles），覆盖租户用户。
// 查不到时回退到 account_id 直查（覆盖 platform admin，他们没 users 行）。
// 两者都未传或查不到时返回 nil（不是错误——普通租户用户本来就没有 platform role）。
func (s *Service) loadPlatformRoles(ctx context.Context, userID, accountID uint) []string {
	if s.platformRp == nil || s.db == nil {
		return nil
	}
	// 路径 1：tenant user
	if userID > 0 {
		if roles, err := s.platformRp.GetRolesByUserID(ctx, userID); err == nil && len(roles) > 0 {
			return roles
		}
	}
	// 路径 2：platform admin（兜底——user_id 查不到时）
	if accountID > 0 {
		if roles, err := s.platformRp.GetRolesByAccountID(ctx, accountID); err == nil {
			return roles
		}
	}
	return nil
}

// Login 租户域登录（业务用户）。需要传 tenant_id；user 必须绑到该 tenant。
//
// 登录安全钩子：
//   - 入口：checkAccountLock 检查是否被锁
//   - 失败：recordFailure 写入 login_attempts，触发窗口计数 + 锁定
//   - 成功：recordSuccess 写入 login_history + 异地告警
func (s *Service) Login(ctx context.Context, req tenantLoginRequest) (*LoginResult, error) {
	// 0. 锁定检查
	if err := s.checkAccountLock(ctx, req.Account); err != nil {
		s.recordFailure(ctx, req.Account, login_security.ScopeTenant, req.TenantID, login_security.FailureAccountLocked)
		return nil, err
	}

	identity, err := ResolveLoginIdentity(ctx, s.db, req.Account, xincontext.NewTenantID(req.TenantID))
	if err != nil {
		switch {
		case errors.Is(err, ErrBackendUnavailable):
			s.recordFailure(ctx, req.Account, login_security.ScopeTenant, req.TenantID, login_security.FailureInvalidPassword)
			return nil, ErrBackendUnavailable
		case errors.Is(err, errAccountNotFound):
			s.recordFailure(ctx, req.Account, login_security.ScopeTenant, req.TenantID, login_security.FailureAccountNotFound)
			return nil, ErrInvalidAccountOrPassword
		case errors.Is(err, errTenantBindingNotFound):
			// 账号存在但未绑该租户：不算密码错误，避免锁定合法账号
			return nil, ErrTenantBindingNotFound
		default:
			s.recordFailure(ctx, req.Account, login_security.ScopeTenant, req.TenantID, login_security.FailureInvalidPassword)
			return nil, ErrInvalidAccountOrPassword
		}
	}

	ok, err := verifyPassword(identity.PasswordHash, req.Password)
	if err != nil || !ok {
		s.recordFailure(ctx, req.Account, login_security.ScopeTenant, req.TenantID, login_security.FailureInvalidPassword)
		return nil, ErrInvalidAccountOrPassword
	}
	if identity.UserStatus != StatusActive {
		s.recordFailure(ctx, req.Account, login_security.ScopeTenant, req.TenantID, login_security.FailureUserDisabled)
		return nil, ErrUserDisabled
	}
	tokens, err := s.generateTokens(ctx, identity.UserID, identity.TenantID, identity.RoleCode, 0)
	if err != nil {
		return nil, err
	}

	// 5. 记录登录成功 + 触发异地告警
	s.recordSuccess(ctx, identity.AccountID, identity.UserID, identity.TenantID,
		login_security.ScopeTenant, tokens.sessionID, identity.RoleCode, nil)

	res := &LoginResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
		Scope:        LoginScopeTenant,
	}
	res.User.ID = uint(identity.UserID)
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

// LoginPrecheck 登录前置检查：验证账号密码后列出可用身份。
//
// 用途：账号可能在多个租户都有 users 记录，或同时是平台账号。
// 前端先调此接口拿到身份列表，让用户选择后再调 /auth/select-tenant
// 或 /auth/platform-login 签发 token。
//
// 错误码：
//   - 账号/密码错 → ErrInvalidAccountOrPassword
//   - 账号无任何身份（无 tenant 身份 + 无 platform 角色）→ ErrNoLoginIdentity
//
// 单身份账号也可以调此接口（precheck 返回 1 个 tenant 身份后，前端
// 直接调 select-tenant 即可），但更直接的做法是跳过 precheck 直接
// 调 /auth/tenant-login。
func (s *Service) LoginPrecheck(ctx context.Context, req loginPrecheckRequest) (*loginPrecheckResult, error) {
	if s.accountRepo == nil || s.platformRp == nil {
		// boot 期依赖未注入的 wiring 错误。理想情况应该在 NewService 里 panic
		// 启动失败,这里仅做请求期兜底:记 Errorf 让 SRE 立刻看到。
		logger.Module("auth").Errorf("LoginPrecheck called with nil deps: accountRepo=%v platformRp=%v", s.accountRepo, s.platformRp)
		return nil, ErrBackendUnavailable
	}

	log := logger.Module("auth")
	// 1. 验证账号密码
	passwordHash, accountID, accountStatus, err := s.accountRepo.GetPasswordAndStatus(ctx, req.Account)
	if err != nil {
		if errors.Is(err, errAccountNotFound) {
			return nil, ErrInvalidAccountOrPassword
		}
		// 其它 DB 故障:记日志(带 account + 原始 err)再返回 1005。
		// 用 %w 包装保留错误链,后续如果加 sentry / otel 可直接 unwrap 拿到根因。
		log.Errorf("LoginPrecheck.GetPasswordAndStatus account=%q: %v", req.Account, err)
		return nil, fmt.Errorf("%w: get password: %w", ErrBackendUnavailable, err)
	}
	if accountStatus != StatusActive {
		return nil, ErrUserDisabled
	}
	ok, err := verifyPassword(passwordHash, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidAccountOrPassword
	}

	// 2. 列出所有 tenant 身份（跨租户，走 RLS bypass）
	tenantIdentities, err := s.accountRepo.ListTenantIdentities(ctx, accountID)
	if err != nil {
		log.Errorf("LoginPrecheck.ListTenantIdentities accountID=%d: %v", accountID, err)
		return nil, fmt.Errorf("%w: list tenant identities: %w", ErrBackendUnavailable, err)
	}

	// 3. 查 platform 角色
	platformRoles, err := s.platformRp.GetRolesByAccountID(ctx, accountID)
	if err != nil {
		log.Errorf("LoginPrecheck.GetRolesByAccountID accountID=%d: %v", accountID, err)
		return nil, fmt.Errorf("%w: get platform roles: %w", ErrBackendUnavailable, err)
	}

	// 4. 业务规则：账号必须至少有 1 个 tenant 身份 或 1 个 platform 角色
	if len(tenantIdentities) == 0 && len(platformRoles) == 0 {
		return nil, ErrNoLoginIdentity
	}

	// 5. 取账号展示信息（real_name / email）从第一个 tenant 身份或留空
	realName := ""
	email := ""
	if len(tenantIdentities) > 0 {
		realName = tenantIdentities[0].RealName
		email = tenantIdentities[0].Email
	}

	return &loginPrecheckResult{
		AccountID:         accountID,
		AccountStatus:     accountStatus,
		RealName:          realName,
		Email:             email,
		PlatformAvailable: len(platformRoles) > 0,
		PlatformRoles:     platformRoles,
		TenantIdentities:  tenantIdentities,
	}, nil
}

// SelectTenant 等价于 Login（tenantLoginRequest）。仅作为语义别名：
// 用于"precheck → 选择身份 → 签发 token"流程，与 /auth/tenant-login
// 共享实现。前端可以从 precheck 返回的 tenant_identities 列表里选一个
// tenant_id，然后调用此接口签发 token。
func (s *Service) SelectTenant(ctx context.Context, req tenantLoginRequest) (*LoginResult, error) {
	return s.Login(ctx, req)
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

	// 0. 锁定检查
	if err := s.checkAccountLock(ctx, req.Account); err != nil {
		s.recordFailure(ctx, req.Account, login_security.ScopePlatform, 0, login_security.FailureAccountLocked)
		return nil, err
	}

	// 1. accounts 表查账号
	passwordHash, accountID, accountStatus, err := s.accountRepo.GetPasswordAndStatus(ctx, req.Account)
	if err != nil {
		if errors.Is(err, errAccountNotFound) {
			s.recordFailure(ctx, req.Account, login_security.ScopePlatform, 0, login_security.FailureAccountNotFound)
			return nil, ErrPlatformLoginAccountNotFound
		}
		s.recordFailure(ctx, req.Account, login_security.ScopePlatform, 0, login_security.FailureInvalidPassword)
		return nil, ErrBackendUnavailable
	}
	if accountStatus != StatusActive {
		s.recordFailure(ctx, req.Account, login_security.ScopePlatform, 0, login_security.FailureUserDisabled)
		return nil, ErrPlatformLoginDisabled
	}

	// 2. 验密码
	ok, err := verifyPassword(passwordHash, req.Password)
	if err != nil || !ok {
		s.recordFailure(ctx, req.Account, login_security.ScopePlatform, 0, login_security.FailureInvalidPassword)
		return nil, ErrInvalidAccountOrPassword
	}

	// 3. 查 platform_roles
	platformRoles, err := s.platformRp.GetRolesByAccountID(ctx, accountID)
	if err != nil {
		return nil, ErrBackendUnavailable
	}
	if len(platformRoles) == 0 {
		// 不是密码错误，不计入失败计数；返回专门错误码让前端识别"非管理员"
		return nil, ErrPlatformLoginNotAdmin
	}

	// 4. 生成 token，tenant_id = 0（platform 域语义）
	//    role 用 "_platform" 占位（不在 tenants / roles 表里）
	//    userID 用 accountID：platform admin 没有 users 行，userID 字段在 JWT 里就是 account_id
	//    （与 LoginResult.User.ID 语义保持一致）
	tokens, err := s.generateTokens(ctx, accountID, 0, "_platform", accountID)
	if err != nil {
		return nil, err
	}

	// 5. 记录登录成功 + 触发异地告警
	s.recordSuccess(ctx, accountID, accountID, 0,
		login_security.ScopePlatform, tokens.sessionID, "_platform", platformRoles)

	res := &LoginResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
		Scope:        LoginScopePlatform,
		User: User{
			ID:            accountID, // 这里 ID 是 account_id（不是 user_id），因为平台用户可能没绑 user
			TenantID:      0,
			Code:          req.Account,
			Role:          RoleCodePlatform,
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

	// 默认沿用 refresh token 里的身份
	targetTenantID := claims.TenantID
	targetUserID := claims.UserID
	targetRole := claims.Role

	// 切租户流程（路径 B 多身份支持）
	if req.TenantID > 0 && req.TenantID != claims.TenantID {
		// 平台 token 不能切租户：platform token 的 UserID 是 account_id，
		// 跟 users.id 是两个空间，没有"切到某个 user 身份"的语义。
		if claims.TenantID == 0 {
			return nil, ErrCrossTenantSwitchFromPlatform
		}

		// 1. 反查 account_id（当前 token 在 claims.TenantID 事务里，
		//    RLS 自动放行该租户的 user 行）
		if s.accountRepo == nil {
			return nil, ErrBackendUnavailable
		}
		accountID, err := s.accountRepo.GetAccountIDByUserID(ctx, claims.UserID)
		if err != nil {
			if errors.Is(err, ErrAccountNotFound) {
				return nil, ErrInvalidRefreshToken
			}
			return nil, ErrBackendUnavailable
		}

		// 2. 跨租户列账号所有身份
		identities, err := s.accountRepo.ListTenantIdentities(ctx, accountID)
		if err != nil {
			return nil, ErrBackendUnavailable
		}

		// 3. 在 identities 里找目标 tenant_id
		var found *pkgauth.TenantIdentity
		for i := range identities {
			if identities[i].TenantID == req.TenantID {
				found = &identities[i]
				break
			}
		}
		if found == nil {
			return nil, ErrTenantBindingNotFound
		}

		targetTenantID = found.TenantID
		targetUserID = found.UserID
		targetRole = found.Role
	}

	// 切租户分支已知 targetUserID 是 users.id（path B 不允许 platform token 切租户，见 L449）。
	// 原地刷新时 targetUserID = claims.UserID：tenant token 时是 users.id；platform token 时是 account_id。
	// 传入 accountID：platform token 时与 userID 等同（让 loadPlatformRoles 走 GetRolesByAccountID 路径）；
	// tenant token 时 accountID 来自反向查询，避免对每个 tenant user 多查一次 account。
	var refreshAccountID uint
	if targetTenantID == 0 && targetUserID > 0 {
		refreshAccountID = targetUserID
	}
	newTokens, err := s.generateTokens(ctx, targetUserID, targetTenantID, targetRole, refreshAccountID)
	if err != nil {
		return nil, err
	}

	// 旧 session 撤销失败不影响新 token 签发 —— 用户已经获得新凭证,
	// 记 warn 让 SRE 留意,不要把可恢复错误升级为请求失败。
	if err := s.session.Revoke(claims.SessionID); err != nil {
		logger.Module("auth").Warnf("revoke old session %q: %v", claims.SessionID, err)
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
			return ErrRegisterFailed
		}
		if t.Status != StatusActive {
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
			INSERT INTO tenant_users (tenant_id, account_id, code, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id`, req.TenantID, newAccount.ID, newUserCode, 1).Scan(&newUserID)
		if err != nil {
			return ErrRegisterFailed
		}

		var roleID uint
		err = querier.QueryRow(ctx, `
			SELECT id FROM tenant_roles
			WHERE is_deleted = FALSE AND tenant_id = $1 AND is_default = TRUE
			LIMIT 1
		`, req.TenantID).Scan(&roleID)
		if err != nil {
			return ErrDefaultRoleNotFound
		}

		_, err = querier.Exec(ctx, `
			INSERT INTO tenant_user_roles (tenant_id, user_id, role_id)
			VALUES ($1, $2, $3)`, req.TenantID, newUserID, roleID)
		if err != nil {
			return ErrRegisterFailed
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(ctx, newUserID, req.TenantID, "user", 0)
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

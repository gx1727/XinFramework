package weixin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/pkg/config"
	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	"gx1727.com/xin/framework/pkg/model"
)

const (
	AuthTypeWxxcx = "wxxcx" // 小程序授权类型
)

type Service struct {
	db              *pgxpool.Pool
	session         SessionManager
	accountAuthRepo model.AccountAuthRepository
	accountRepo     model.AccountRepository
	tenantRepo      model.TenantRepository
	roleRepo        model.RoleRepository
	userRepo        model.UserRepository
}

type SessionManager interface {
	Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
	Validate(sessionID string) (bool, error)
	Revoke(sessionID string) error
}

func NewService(
	db *pgxpool.Pool,
	session SessionManager,
	accountAuthRepo model.AccountAuthRepository,
	accountRepo model.AccountRepository,
	tenantRepo model.TenantRepository,
	roleRepo model.RoleRepository,
	userRepo model.UserRepository,
) *Service {
	return &Service{
		db:              db,
		session:         session,
		accountAuthRepo: accountAuthRepo,
		accountRepo:     accountRepo,
		tenantRepo:      tenantRepo,
		roleRepo:        roleRepo,
		userRepo:        userRepo,
	}
}

// Code2Session 调用微信接口获取 openid 和 session_key
func (s *Service) Code2Session(ctx context.Context, code string) (*Code2SessionResponse, error) {
	if Cfg().AppID == "" || Cfg().AppSecret == "" {
		return nil, ErrBackendUnavailable
	}

	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		Cfg().AppID,
		Cfg().AppSecret,
		code,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, ErrWeChatAPIFailed
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrWeChatAPIFailed
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrWeChatAPIFailed
	}

	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, ErrWeChatAPIFailed
	}

	if result.ErrCode != 0 {
		if result.ErrCode == 40029 || result.ErrCode == 40125 {
			return nil, ErrInvalidCode
		}
		return nil, fmt.Errorf("%w: %s", ErrWeChatAPIFailed, result.ErrMsg)
	}

	return &Code2SessionResponse{
		OpenID:     result.OpenID,
		SessionKey: result.SessionKey,
		UnionID:    result.UnionID,
	}, nil
}

// GetPhoneNumber 获取用户手机号
func (s *Service) GetPhoneNumber(ctx context.Context, code string) (*PhoneNumberResponse, error) {
	if Cfg().AppID == "" || Cfg().AppSecret == "" {
		return nil, ErrBackendUnavailable
	}

	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		accessToken,
	)

	reqBody := fmt.Sprintf(`{"code":"%s"}`, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, ErrWeChatAPIFailed
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrWeChatAPIFailed
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrWeChatAPIFailed
	}

	var result struct {
		ErrCode   int        `json:"errcode"`
		ErrMsg    string     `json:"errmsg"`
		PhoneInfo *PhoneInfo `json:"phone_info,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, ErrWeChatAPIFailed
	}

	if result.ErrCode != 0 {
		if result.ErrCode == 40029 || result.ErrCode == 40125 {
			return nil, ErrPhoneCodeInvalid
		}
		return nil, fmt.Errorf("%w: %s", ErrWeChatAPIFailed, result.ErrMsg)
	}

	return &PhoneNumberResponse{
		PhoneInfo: *result.PhoneInfo,
	}, nil
}

// getAccessToken 获取 access_token
func (s *Service) getAccessToken(ctx context.Context) (string, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		Cfg().AppID,
		Cfg().AppSecret,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", ErrWeChatAPIFailed
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ErrWeChatAPIFailed
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ErrWeChatAPIFailed
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", ErrWeChatAPIFailed
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("%w: %s", ErrWeChatAPIFailed, result.ErrMsg)
	}

	return result.AccessToken, nil
}

// LoginByWeChat 小程序登录
func (s *Service) LoginByWeChat(ctx context.Context, code string) (*LoginResult, error) {
	if s.db == nil || s.session == nil {
		return nil, ErrBackendUnavailable
	}

	sessionResp, err := s.Code2Session(ctx, code)
	if err != nil {
		return nil, err
	}

	openID := sessionResp.OpenID

	// 获取或创建默认租户
	tenant, err := s.getOrCreateDefaultTenant(ctx)
	if err != nil {
		return nil, err
	}
	tenantID := tenant.ID

	// 查找已有的微信授权记录 - 使用事务查询以满足RLS策略
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID)))
	if err != nil {
		return nil, err
	}

	var userID uint
	var userCode string
	var userStatus int16
	var roleCode string
	isNewUser := false

	// 查询微信授权记录
	var existingAuthID uint
	var existingAccountID uint
	err = tx.QueryRow(ctx, `
		SELECT id, account_id
		FROM account_auths
		WHERE is_deleted = FALSE AND tenant_id = $1 AND type = $2 AND openid = $3
		LIMIT 1
	`, tenantID, AuthTypeWxxcx, openID).Scan(&existingAuthID, &existingAccountID)

	if errors.Is(err, pgx.ErrNoRows) || err != nil {
		// 新用户：创建账号、用户、授权记录
		isNewUser = true
		accountID, newUserID, newUserCode, err := s.createWeChatUser(ctx, tenantID, openID, sessionResp.UnionID, sessionResp.SessionKey)
		if err != nil {
			return nil, err
		}
		userID = newUserID
		userCode = newUserCode
		roleCode = "user"
		userStatus = 1

		_ = accountID // unused
	} else {
		// 老用户：查询用户信息
		err = tx.QueryRow(ctx, `
			SELECT u.id, u.tenant_id, u.code, u.status
			FROM users u
			WHERE u.account_id = $1 AND u.is_deleted = FALSE
			LIMIT 1
		`, existingAccountID).Scan(&userID, &tenantID, &userCode, &userStatus)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		if userStatus != 1 {
			return nil, ErrUserDisabled
		}

		// 更新 session_key
		_, err = tx.Exec(ctx, `
			UPDATE account_auths SET session_key = $1, updated_at = NOW()
			WHERE is_deleted = FALSE AND id = $2
		`, sessionResp.SessionKey, existingAuthID)
		if err != nil {
			return nil, err
		}

		// 获取角色
		err = tx.QueryRow(ctx, `
			SELECT r.code
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1
			ORDER BY ur.id ASC LIMIT 1
		`, userID).Scan(&roleCode)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		}
	}

	// 提交查询事务（只读操作）
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(ctx, userID, tenantID, roleCode)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:        tokens.accessToken,
		RefreshToken: tokens.refreshToken,
		User: LoginUser{
			ID:       userID,
			OpenID:   openID,
			UnionID:  sessionResp.UnionID,
			TenantID: tenantID,
			Code:     userCode,
			Role:     roleCode,
			Status:   userStatus,
		},
		IsNewUser: isNewUser,
	}, nil
}

func (s *Service) getOrCreateDefaultTenant(ctx context.Context) (*model.Tenant, error) {
	var tenant model.Tenant
	err := s.db.QueryRow(ctx, `
		SELECT id, code, name, status
		FROM tenants
		WHERE code = 'default' AND is_deleted = FALSE
		LIMIT 1
	`).Scan(&tenant.ID, &tenant.Code, &tenant.Name, &tenant.Status)
	if err == nil {
		return &tenant, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var newTenantID uint
	err = s.db.QueryRow(ctx, `
		INSERT INTO tenants (code, name, status, created_at, updated_at)
		VALUES ('default', 'Default Tenant', 1, NOW(), NOW())
		RETURNING id
	`).Scan(&newTenantID)
	if err != nil {
		return nil, err
	}

	tenant.ID = newTenantID
	tenant.Code = "default"
	tenant.Name = "Default Tenant"
	tenant.Status = 1
	return &tenant, nil
}

func (s *Service) createWeChatUser(ctx context.Context, tenantID uint, openID, unionID, sessionKey string) (uint, uint, string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, 0, "", err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", strconv.Itoa(int(tenantID)))
	if err != nil {
		return 0, 0, "", err
	}

	// 创建账号
	var accountID uint
	err = tx.QueryRow(ctx, `
		INSERT INTO accounts (username, status, created_at, updated_at)
		VALUES ($1, 1, NOW(), NOW())
		RETURNING id
	`, openID).Scan(&accountID)
	if err != nil {
		return 0, 0, "", err
	}

	// 创建用户
	userCode, err := s.generateUserCode(ctx, tx, tenantID)
	if err != nil {
		return 0, 0, "", err
	}

	var userID uint
	err = tx.QueryRow(ctx, `
		INSERT INTO users (tenant_id, account_id, code, status, created_at, updated_at)
		VALUES ($1, $2, $3, 1, NOW(), NOW())
		RETURNING id
	`, tenantID, accountID, userCode).Scan(&userID)
	if err != nil {
		return 0, 0, "", err
	}

	// 获取默认角色
	var roleID uint
	err = tx.QueryRow(ctx, `
		SELECT id FROM roles
		WHERE is_deleted = FALSE AND tenant_id = $1 AND is_default = TRUE
		LIMIT 1
	`, tenantID).Scan(&roleID)
	if err != nil {
		return 0, 0, "", errors.New("default role not found")
	}

	// 分配角色
	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (tenant_id, user_id, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, tenantID, userID, roleID)
	if err != nil {
		return 0, 0, "", err
	}

	// 创建微信授权记录
	_, err = tx.Exec(ctx, `
		INSERT INTO account_auths (tenant_id, account_id, type, openid, unionid, session_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`, tenantID, accountID, AuthTypeWxxcx, openID, unionID, sessionKey)
	if err != nil {
		return 0, 0, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, "", err
	}

	return accountID, userID, userCode, nil
}

func (s *Service) generateUserCode(ctx context.Context, tx pgx.Tx, tenantID uint) (string, error) {
	var seq int64
	err := tx.QueryRow(ctx, `
		INSERT INTO tenant_user_seq (tenant_id, seq)
		VALUES ($1, 1)
		ON CONFLICT (tenant_id) DO UPDATE SET seq = tenant_user_seq.seq + 1
		RETURNING seq
	`, tenantID).Scan(&seq)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("U%03d%06d", tenantID, seq), nil
}

type tokenPair struct {
	accessToken  string
	refreshToken string
}

func (s *Service) generateTokens(ctx context.Context, userID, tenantID uint, role string) (*tokenPair, error) {
	sessionID := uuid.NewString()
	refreshTTL := time.Duration(config.Get().JWT.RefreshExpire) * time.Second

	if err := s.session.Create(sessionID, userID, tenantID, role, refreshTTL); err != nil {
		return nil, err
	}

	accessToken, err := jwtpkg.Generate(&config.Get().JWT, userID, tenantID, role, sessionID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwtpkg.GenerateWithType(&config.Get().JWT, userID, tenantID, role, sessionID, jwtpkg.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	return &tokenPair{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

// UpdatePhoneByWeChat 通过微信更新用户手机号
func (s *Service) UpdatePhoneByWeChat(ctx context.Context, userID uint, code string) (string, error) {
	phoneResp, err := s.GetPhoneNumber(ctx, code)
	if err != nil {
		return "", err
	}

	if phoneResp.PhoneInfo.PhoneNumber == "" {
		return "", ErrInvalidPhoneNumber
	}

	phone := phoneResp.PhoneInfo.PurePhoneNumber
	if phone == "" {
		phone = phoneResp.PhoneInfo.PhoneNumber
	}

	err = s.userRepo.UpdatePhone(ctx, userID, phone)
	if err != nil {
		return "", err
	}

	return phone, nil
}

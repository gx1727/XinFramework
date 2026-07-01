// Package jwt 提供 JWT token 的签发、解析与 Claims 类型。
//
// 本包仅负责“令牌本身”的生成与校验，会话有效性由 session.SessionManager
// 独立检查。Auth 中间件在中间件侧会把这两步串联起来。
//
// 关键概念：
//   - TokenTypeAccess / TokenTypeRefresh：区分访问令牌与刷新令牌
//   - SysRoleSuperAdmin：sys 超级管理员角色编码，跨租户特权
//   - Claims.SysRoles：登录时查询 sys_user_roles → sys_roles.code 填入，
//     中间件通过 jwt.SysRoleSuperAdmin 常量识别并短路放行
//
// 使用流程：
//  1. 登录成功 → GenerateWithSysRoles 签发 access + refresh 两个 token
//  2. 客户端调用 API 时带 Authorization: Bearer <access>
//  3. Auth 中间件调 Validate 解析 + session.Manager().Validate(SessionID) 检查存活
//  4. 接近过期时调 refresh 接口重签发
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gx1727.com/xin/framework/pkg/config"
)

// Token 类型常量。
const (
	TokenTypeAccess  = "access"  // 访问令牌
	TokenTypeRefresh = "refresh" // 刷新令牌
)

// SysRoleSuperAdmin sys 超级管理员角色编码。
// 拥有此角色的账号可以无视租户边界与 DataScope，中间件 requireWithSpecs 会
// 检到 SysRoles 含本常量后直接放行。
const SysRoleSuperAdmin = "super_admin"

// Claims 是框架自定义的 JWT 载荷。
// UserID / TenantID / Role 在 Auth 中间件被取出注入到 XinContext；
// SessionID 供 session.Manager().Validate() 做存活检查；
// SysRoles 供 requireWithSpecs 判断 super_admin 短路。
//
// ImpersonatedBy / ImpersonationSessionID 用于 super_admin「模拟登录租户」：
//   - ImpersonatedBy > 0 表示当前 token 是模拟签发的，原账号（super_admin）= ImpersonatedBy
//   - 此时 Claims.SysRoles 应为空，模拟期间走租户 RBAC，不享受 sys 域短路
//   - ImpersonationSessionID 记录原 sys 会话 ID；前端调 /auth/refresh（不传 tenant_id）
//     即可恢复原 sys token，实现「退出模拟」
type Claims struct {
	UserID                 uint     `json:"user_id"`
	TenantID               uint     `json:"tenant_id"`
	Role                   string   `json:"role"`
	SessionID              string   `json:"sid"`
	TokenType              string   `json:"token_type"`
	SysRoles               []string `json:"sys_role_codes,omitempty"`
	ImpersonatedBy         uint     `json:"imp_by,omitempty"`
	ImpersonationSessionID string   `json:"imp_sid,omitempty"`
	jwt.RegisteredClaims
}

// Generate 签发默认的访问令牌（TokenTypeAccess）。
func Generate(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string) (string, error) {
	return GenerateWithType(cfg, userID, tenantID, role, sessionID, TokenTypeAccess)
}

// GenerateWithType 签发指定类型的令牌（access 或 refresh）。
func GenerateWithType(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, tokenType string) (string, error) {
	return GenerateWithSysRoles(cfg, userID, tenantID, role, sessionID, nil, tokenType)
}

// GenerateWithSysRoles 签发携带 sys 角色信息的令牌。
// sysRoles 可为 nil；非空时 JWT 中会带上这些角色码，
// 中间件据此识别 super_admin 并短路放行。
func GenerateWithSysRoles(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, sysRoles []string, tokenType string) (string, error) {
	expire := cfg.Expire
	if tokenType == TokenTypeRefresh {
		expire = cfg.RefreshExpire
	}

	claims := Claims{
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		SessionID: sessionID,
		TokenType: tokenType,
		SysRoles:  sysRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// GenerateImpersonation 签发 super_admin 模拟登录租户的令牌。
//
// 参数：
//   - adminUserID：租户内被模拟的用户 ID（一般取 admin role 绑定的 user）
//   - tenantID：目标租户
//   - sessionID：新创建的 auth_sessions.id（按 account 维度复用）
//   - impersonatedBy：原 super_admin 的 account_id（= 当前 Context.UserID）
//   - impersonationSID：原 sys 会话 ID；前端"退出模拟"时调 /auth/refresh
//     用 refresh_token 不带 tenant_id，刷新得到的新 token 仍然是原 sys 会话
//   - sysRoles：必须为 nil；模拟期间不走 super_admin 短路
func GenerateImpersonation(cfg *config.JWTConfig, adminUserID, tenantID uint, role, sessionID string, impersonatedBy uint, impersonationSID string, tokenType string) (string, error) {
	expire := cfg.Expire
	if tokenType == TokenTypeRefresh {
		expire = cfg.RefreshExpire
	}
	claims := Claims{
		UserID:                 adminUserID,
		TenantID:               tenantID,
		Role:                   role,
		SessionID:              sessionID,
		TokenType:              tokenType,
		ImpersonatedBy:         impersonatedBy,
		ImpersonationSessionID: impersonationSID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// IsImpersonation 判断当前 token 是否为模拟签发的。
func (c *Claims) IsImpersonation() bool {
	return c != nil && c.ImpersonatedBy > 0
}

// HasSysRole 便捷判断：claims 是否包含指定的 sys 级角色。
// 典型调用：HasSysRole(SysRoleSuperAdmin)。
func (c *Claims) HasSysRole(role string) bool {
	if c == nil || role == "" {
		return false
	}
	for _, r := range c.SysRoles {
		if r == role {
			return true
		}
	}
	return false
}

func Validate(tokenString string, cfg *config.JWTConfig) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ValidateRefresh(tokenString string, cfg *config.JWTConfig) (*Claims, error) {
	claims, err := Validate(tokenString, cfg)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}
	return claims, nil
}

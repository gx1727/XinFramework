package auth

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrInvalidAccountOrPassword = resp.Err(1001, "账号或密码错误")
	ErrInvalidToken             = resp.Err(1002, "无效的Token")
	ErrInvalidRefreshToken      = resp.Err(1014, "无效的刷新令牌")
	ErrTenantRequired           = resp.Err(1015, "tenant_id 必填")

	ErrUserDisabled          = resp.Err(1003, "账号已被禁用")
	ErrTenantBindingNotFound = resp.Err(1004, "用户未绑定任何租户")

	ErrAccountAlreadyExists = resp.Err(1010, "账号已存在")
	ErrTenantNotFound       = resp.Err(1011, "租户不存在或已被禁用")
	ErrDefaultRoleNotFound  = resp.Err(1012, "未找到默认角色")
	ErrRegisterFailed       = resp.Err(1013, "注册失败")

	ErrBackendUnavailable  = resp.Err(1005, "服务后端未初始化或不可用")
	ErrSessionCreateFailed = resp.Err(1006, "创建会话失败")
	ErrSessionRevokeFailed = resp.Err(1007, "注销会话失败")
	ErrGenerateTokenFailed = resp.Err(1008, "生成令牌失败")
	ErrAccountNotFound     = resp.Err(1009, "账号不存在")
)

var (
	ErrInvalidHashFormat = errors.New("invalid argon2id hash format")
	ErrUnsupportedHash   = errors.New("unsupported hash algorithm")
)

var (
	errAccountNotFound       = errors.New("account not found")
	errTenantBindingNotFound = errors.New("tenant binding not found")
)

// 登录安全策略（锁定 / 异地）错误（错误码段 1020-1029）。
//
// 包含 ErrAccountLocked、ErrTooManyAttempts、ErrAnomalyLoginConfirmed 三种语义。
// 详见 framework/pkg/login_security 包。
var (
	ErrAccountLocked           = resp.Err(1020, "账号已被锁定，请在锁定结束后再试或联系管理员")
	ErrTooManyAttempts          = resp.Err(1021, "登录失败次数过多，请稍后再试")
	ErrAnomalyLoginConfirmed    = resp.Err(1022, "检测到异地登录，如非本人操作请尽快修改密码")
	ErrLoginSecurityUnavailable = resp.Err(1023, "登录安全服务暂时不可用")
)

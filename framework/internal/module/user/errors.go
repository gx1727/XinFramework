package user

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrInvalidAccountOrPassword = resp.NewError(1001, "账号或密码错误")
	ErrInvalidToken             = resp.NewError(1002, "无效的Token")
	ErrInvalidRefreshToken      = resp.NewError(1014, "无效的刷新令牌")

	ErrUserDisabled          = resp.NewError(1003, "账号已被禁用")
	ErrTenantBindingNotFound = resp.NewError(1004, "用户未绑定任何租户")

	ErrAccountAlreadyExists = resp.NewError(1010, "账号已存在")
	ErrTenantNotFound       = resp.NewError(1011, "租户不存在或已被禁用")
	ErrDefaultRoleNotFound  = resp.NewError(1012, "未找到默认角色")
	ErrRegisterFailed       = resp.NewError(1013, "注册失败")

	ErrBackendUnavailable  = resp.NewError(1005, "服务后端未初始化或不可用")
	ErrSessionCreateFailed = resp.NewError(1006, "创建会话失败")
	ErrSessionRevokeFailed = resp.NewError(1007, "注销会话失败")
	ErrGenerateTokenFailed = resp.NewError(1008, "生成令牌失败")
	ErrAccountNotFound     = resp.NewError(1009, "账号不存在")
)

var (
	ErrInvalidHashFormat = errors.New("invalid argon2id hash format")
	ErrUnsupportedHash   = errors.New("unsupported hash algorithm")
)

var (
	errAccountNotFound       = errors.New("account not found")
	errTenantBindingNotFound = errors.New("tenant binding not found")
)

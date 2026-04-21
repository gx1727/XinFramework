package auth

import "gx1727.com/xin/pkg/resp"

// Auth 模块错误码分配：1000 - 1999
var (
	// HTTP 401 系列
	ErrInvalidAccountOrPassword = resp.NewError(401, 1001, "账号或密码错误")
	ErrInvalidToken             = resp.NewError(401, 1002, "无效的 Token")

	// HTTP 403 系列
	ErrUserDisabled          = resp.NewError(403, 1003, "账号已被禁用")
	ErrTenantBindingNotFound = resp.NewError(403, 1004, "用户未绑定任何租户")

	// HTTP 500 系列
	ErrBackendUnavailable  = resp.NewError(500, 1005, "服务后端未初始化或不可用")
	ErrSessionCreateFailed = resp.NewError(500, 1006, "创建会话失败")
	ErrSessionRevokeFailed = resp.NewError(500, 1007, "注销会话失败")
	ErrGenerateTokenFailed = resp.NewError(500, 1008, "生成令牌失败")
)

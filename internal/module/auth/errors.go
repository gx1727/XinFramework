package auth

import "gx1727.com/xin/pkg/resp"

// Auth 模块错误码分配：1000 - 1999
// 统一使用 HTTP 200，HTTP 语义全由业务 Code 表达
var (
	// 账号密码相关
	ErrInvalidAccountOrPassword = resp.NewError(1001, "账号或密码错误")
	ErrInvalidToken             = resp.NewError(1002, "无效的 Token")

	// 权限相关
	ErrUserDisabled          = resp.NewError(1003, "账号已被禁用")
	ErrTenantBindingNotFound = resp.NewError(1004, "用户未绑定任何租户")

	// 系统错误
	ErrBackendUnavailable  = resp.NewError(1005, "服务后端未初始化或不可用")
	ErrSessionCreateFailed = resp.NewError(1006, "创建会话失败")
	ErrSessionRevokeFailed = resp.NewError(1007, "注销会话失败")
	ErrGenerateTokenFailed = resp.NewError(1008, "生成令牌失败")
)

package weixin

import (
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrWeChatAPIFailed    = resp.NewError(4001, "wechat api call failed")
	ErrInvalidCode        = resp.NewError(4002, "invalid wechat code")
	ErrSessionKeyExpired  = resp.NewError(4003, "session key expired")
	ErrPhoneCodeInvalid   = resp.NewError(4004, "invalid phone code")
	ErrBackendUnavailable = resp.NewError(5001, "backend service unavailable")
	ErrUserDisabled       = resp.NewError(4005, "user is disabled")
	ErrInvalidPhoneNumber = resp.NewError(4006, "invalid phone number")
)

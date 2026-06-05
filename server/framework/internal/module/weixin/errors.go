package weixin

import (
	"gx1727.com/xin/framework/pkg/resp"
)

var (
	ErrWeChatAPIFailed    = resp.Err(12001, "wechat api call failed")
	ErrInvalidCode        = resp.Err(12002, "invalid wechat code")
	ErrSessionKeyExpired  = resp.Err(12003, "session key expired")
	ErrPhoneCodeInvalid   = resp.Err(12004, "invalid phone code")
	ErrBackendUnavailable = resp.Err(12005, "backend service unavailable")
	ErrUserDisabled       = resp.Err(12006, "user is disabled")
	ErrInvalidPhoneNumber = resp.Err(12007, "invalid phone number")
)

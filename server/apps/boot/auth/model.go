package auth

import (
	"context"
	"errors"
	"time"

	pkgauth "gx1727.com/xin/framework/pkg/auth"
)

// 类型别名（type alias）让 apps/boot/auth 与 framework 的 user/weixin
// 模块共享同一个 struct 类型，避免 Phase 2 之后的类型不兼容。
//
// Account / AccountAuth / AccountRepository / AccountAuthRepository 的
// 字段集与方法签名在 apps/boot/auth 与 framework/pkg/auth 完全对齐；
// apps/boot/auth 是实现方，framework/pkg/auth 是公开契约。
type Account = pkgauth.Account
type AccountAuth = pkgauth.AccountAuth
type AccountRepository = pkgauth.AccountRepository
type AccountAuthRepository = pkgauth.AccountAuthRepository

// apps/boot/auth-local errors (preserved here so call sites that used
// to import them keep compiling).
var (
	ErrAccountNotFoundDB      = errors.New("account not found")
	ErrAccountAlreadyExistsDB = errors.New("account already exists")
	ErrAccountAuthNotFoundDB  = errors.New("account auth not found")
)

// Compile-time guard: keeps the time import in case future fields are
// added here. Harmless if unused.
var _ = time.Time{}
var _ context.Context
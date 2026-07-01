// Package auth 的公开错误哨兵。
//
// 业务模块（如 apps/sys/user 在新建可登录账号时检查 phone/email
// 是否已被占用）只能 import framework/* 包，访问不到 apps/boot/auth 私有
// 的 errAccountNotFound。因此在 framework 这边再声明一份同名公开错误。
//
// 命名 / 错误码必须与 apps/boot/auth/errors.go 中的保持一致，以便反射
// 类（errors.Is）跨包识别。
package auth

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

// 公开错误哨兵
var (
	// ErrAccountNotFound 账号不存在。
	// apps/boot/auth.PostgresAccountRepository 在 row not found 时返回此错误。
	// 业务模块（通过 framework/pkg/auth.AccountRepository 接口）拿到的也是这个错误。
	ErrAccountNotFound = resp.Err(1009, "账号不存在")
)

// errAccountNotFoundUnwrapped 是 framework 层的内部未包装哨兵，
// 主要用于仓库实现侧对 pgx.ErrNoRows 的回退包装。业务模块不需要直接
// 引用该变量。
var errAccountNotFoundUnwrapped = errors.New("account not found")

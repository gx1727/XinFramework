// Package dict 数据字典
package dict

import "gx1727.com/xin/framework/pkg/resp"

var (
	// ErrDictNotFound 字典不存在
	ErrDictNotFound = resp.Err(10001, "字典不存在")

	// ErrDictCodeExists 字典编码已存在（同租户内）
	ErrDictCodeExists = resp.Err(10002, "字典编码已存在")

	// ErrDictItemNotFound 字典项不存在
	ErrDictItemNotFound = resp.Err(10003, "字典项不存在")

	// ErrDictItemCodeExists 字典项编码已存在（同字典内）
	ErrDictItemCodeExists = resp.Err(10004, "字典项编码已存在")

	// ErrDictHasItems 字典下仍有字典项，禁止删除
	ErrDictHasItems = resp.Err(10005, "字典下仍有字典项，无法删除")

	// ============ Phase 0022: platform / tenant 分层错误码 ============

	// ErrPlatformDictImmutable 租户尝试改/删平台字典
	ErrPlatformDictImmutable = resp.Err(10006, "平台字典不可由租户修改")

	// ErrDictInvisible 字典对当前租户不可见（visibility=invisible 或 whitelist 未命中）
	ErrDictInvisible = resp.Err(10007, "字典对当前租户不可见")

	// ErrDictReadonly 字典对当前租户只读，不允许覆盖字典项
	ErrDictReadonly = resp.Err(10008, "字典为只读，不可覆盖字典项")

	// ErrPlatformItemHasOverrides 删除 platform item 时仍有租户覆盖
	ErrPlatformItemHasOverrides = resp.Err(10009, "该字典项存在租户覆盖，无法删除")

	// ErrPlatformItemMismatch 覆盖项与 platform item 不匹配
	ErrPlatformItemMismatch = resp.Err(10010, "覆盖项与平台字典项不匹配")

	// ErrInvalidAccess access 取值非法
	ErrInvalidAccess = resp.Err(10011, "访问级别取值非法")

	// ErrInvalidScope scope 取值非法
	ErrInvalidScope = resp.Err(10012, "scope 取值非法")

	// ErrInvalidVisibility visibility 取值非法
	ErrInvalidVisibility = resp.Err(10013, "visibility 取值非法")

	// ErrResolveFailed 解析合并字典失败
	ErrResolveFailed = resp.Err(10014, "解析字典失败")
)
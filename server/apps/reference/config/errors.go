// Package config 通用配置 - 错误定义
//
// 错误码段：18xxx（与 dict 的 10xxx、platform_menu 的 15xxx、platform_tenant 的 16xxx 不冲突）
//
// 见 framework/pkg/resp/errors.go 的分段约定。
package config

import "gx1727.com/xin/framework/pkg/resp"

const (
	CodeConfig = 18000
)

// ============ 基础业务错误 ============

var (
	ErrGroupNotFound       = resp.Err(18001, "配置分组不存在")
	ErrGroupCodeExists     = resp.Err(18002, "配置分组编码已存在")
	ErrGroupHasItems       = resp.Err(18003, "配置分组下仍有配置项，无法删除")
	ErrGroupIsSystem       = resp.Err(18004, "系统分组不可修改")

	ErrItemNotFound        = resp.Err(18005, "配置项不存在")
	ErrItemKeyExists       = resp.Err(18006, "配置项 key 已存在")
	ErrItemIsReadonly      = resp.Err(18007, "配置项为只读，无法修改")
	ErrItemIsSystem        = resp.Err(18008, "系统配置项不可修改")
	ErrInvalidItemType     = resp.Err(18009, "配置项 type 取值非法")
	ErrInvalidValueForType = resp.Err(18010, "配置项 value 与 type 不匹配")
	ErrValueNotInOptions   = resp.Err(18011, "配置项 value 不在 options 范围内")
)

// ============ Platform / Tenant 分层错误（与 dict 对齐）============
//
// 租户尝试改/删平台 group 或 platform item 时的拒绝错误。

var (
	// ErrPlatformGroupImmutable 租户尝试改/删 platform group
	ErrPlatformGroupImmutable = resp.Err(18012, "平台配置分组不可由租户修改")

	// ErrPlatformItemHasOverrides 删除 platform item 但仍有租户覆盖
	ErrPlatformItemHasOverrides = resp.Err(18013, "该配置项存在租户覆盖，无法删除")

	// ErrPlatformItemMismatch override 引用不存在的 platform item
	ErrPlatformItemMismatch = resp.Err(18014, "override 引用不存在的 platform item")

	// ErrGroupInvisible 平台 group 对当前租户不可见
	ErrGroupInvisible = resp.Err(18015, "配置分组对当前租户不可见")

	// ErrGroupReadonly 平台 group 对当前租户只读（visibility=readonly 或 whitelist 黑名单）
	ErrGroupReadonly = resp.Err(18016, "配置分组为只读，不可覆盖配置项")

	// ErrInvalidAccess access 取值非法
	ErrInvalidAccess = resp.Err(18017, "访问级别取值非法（应为 invisible / readonly / editable）")

	// ErrInvalidVisibility visibility 取值非法
	ErrInvalidVisibility = resp.Err(18018, "可见性策略取值非法（应为 all / whitelist / blacklist）")

	// ErrResolveFailed 解析合并配置失败
	ErrResolveFailed = resp.Err(18019, "解析配置失败")
)

// ============ 旧错误码映射（向后兼容标记，可选保留）============
//
// 以下变量名保留以兼容旧代码引用，但实际值已经是 resp.Err 类型。
//
// 不再提供 mapRepoError() 转换函数——所有调用点已改为直接返回 resp.Err。

// keep reference for tests; actual values are resp.Err now
var _ = []error{
	ErrGroupNotFound, ErrGroupCodeExists, ErrGroupHasItems, ErrGroupIsSystem,
	ErrItemNotFound, ErrItemKeyExists, ErrItemIsReadonly, ErrItemIsSystem,
	ErrInvalidItemType, ErrInvalidValueForType, ErrValueNotInOptions,
}

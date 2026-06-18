// Package config 通用配置 - 错误定义
package config

import "errors"

// 业务错误（service 层返回，handler 层识别后转 HTTP 状态码）
var (
	ErrGroupNotFound       = errors.New("config: group not found")
	ErrGroupCodeExists     = errors.New("config: group code already exists")
	ErrGroupHasItems       = errors.New("config: group has items")
	ErrGroupIsSystem       = errors.New("config: group is system-protected")

	ErrItemNotFound        = errors.New("config: item not found")
	ErrItemKeyExists       = errors.New("config: item key already exists")
	ErrItemIsReadonly      = errors.New("config: item is read-only")
	ErrItemIsSystem        = errors.New("config: item is system-protected")
	ErrInvalidItemType     = errors.New("config: invalid item type")
	ErrInvalidValueForType = errors.New("config: value does not match type")
	ErrValueNotInOptions   = errors.New("config: value not in options")
)

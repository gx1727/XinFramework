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
)

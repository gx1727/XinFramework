package sysorg

import (
	"errors"

	"gx1727.com/xin/framework/pkg/resp"
)

const (
	CodeSysOrg = 15500
)

var (
	ErrSysOrgNotFound      = resp.Err(15501, "平台组织不存在")
	ErrBackendUnavailable  = resp.Err(15599, "服务后端未初始化或不可用")
)

var (
	errSysOrgNotFoundDB = errors.New("sys_org not found in db")
)

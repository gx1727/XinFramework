// Package message 提供租户域站内信（in-site messaging）能力。
//
// 位置：apps/tenant/message（租户基础设施，与 user/role/menu 同档）
// 数据域：tenant（tenant_messages 表 + RLS，tenant_id=0 短路口子用于平台广播）
// 错误码段：16001-16999（详见 framework/pkg/resp/errors.go）
// 资源码：message（详见 framework/pkg/permission/constants.go）
package message

import "gx1727.com/xin/framework/pkg/resp"

var (
	ErrMessageNotFound   = resp.Err(16001, "站内信不存在")
	ErrRecipientEmpty    = resp.Err(16002, "收件人不能为空")
	ErrSubjectEmpty      = resp.Err(16003, "主题不能为空")
	ErrSenderMismatch    = resp.Err(16004, "只能操作自己发送的信件")
	ErrRecipientMismatch = resp.Err(16005, "只能操作自己收到的信件")
	ErrMessageTypeInvalid = resp.Err(16006, "无效的消息类型")
	ErrPriorityInvalid    = resp.Err(16007, "无效的优先级")
)
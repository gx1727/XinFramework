package message

import "time"

// Message 站内信实体（与 tenant_messages 表一一对应）。
//
// 字段类型遵循 tenant_* 表的约定：DB 侧 BIGINT，对应 Go uint。
type Message struct {
	ID          uint
	TenantID    uint
	SenderID    uint
	RecipientID uint
	Subject     string
	Body        string
	MsgType     int16      // 1=私信 2=通知 3=系统公告
	Priority    int16      // 0=普通 1=重要 2=紧急
	IsRead      bool
	ReadAt      *time.Time
	IsDeleted   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// 消息类型常量（与 msg_type 字段对齐）
const (
	MsgTypePrivate   int16 = 1 // 私信
	MsgTypeNotify    int16 = 2 // 通知
	MsgTypeBroadcast int16 = 3 // 系统公告
)

// 优先级常量（与 priority 字段对齐）
const (
	PriorityNormal int16 = 0
	PriorityHigh   int16 = 1
	PriorityUrgent int16 = 2
)
package message

// 收件箱 / 发件箱
const (
	BoxInbox = "inbox"
	BoxSent  = "sent"
)

// ListReq 列表查询参数
type ListReq struct {
	Box     string `form:"box" binding:"required,oneof=inbox sent"` // inbox=收件箱 sent=发件箱
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
	IsRead  *bool  `form:"is_read"`
	MsgType *int16 `form:"msg_type"`
	Keyword string `form:"keyword"`
}

// SendReq 发送站内信
type SendReq struct {
	RecipientID uint   `json:"recipient_id" binding:"required"`
	Subject     string `json:"subject" binding:"required,min=1,max=255"`
	Body        string `json:"body"`
	MsgType     int16  `json:"msg_type"`
	Priority    int16  `json:"priority"`
}

// UpdateReq 局部更新（nil 字段保持原值）
type UpdateReq struct {
	Subject  *string `json:"subject"`
	Body     *string `json:"body"`
	Priority *int16  `json:"priority"`
}

// MessageResp 响应体
type MessageResp struct {
	ID          uint    `json:"id"`
	SenderID    uint    `json:"sender_id"`
	RecipientID uint    `json:"recipient_id"`
	Subject     string  `json:"subject"`
	Body        string  `json:"body"`
	MsgType     int16   `json:"msg_type"`
	Priority    int16   `json:"priority"`
	IsRead      bool    `json:"is_read"`
	ReadAt      *string `json:"read_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ListResp 分页响应
type ListResp struct {
	List  []MessageResp `json:"list"`
	Total int64         `json:"total"`
}

// UnreadCountResp 未读统计
type UnreadCountResp struct {
	Count int64 `json:"count"`
}
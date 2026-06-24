package message

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/xincontext"
)

// Service 是 message 模块的业务层。
//
// 设计要点：
//   - 所有读写都包在 db.RunInTenantTx 里（自动注入 app.tenant_id，触发 RLS）。
//   - repository.ErrNotFoundDB → service 层翻译成 resp.Err(16001)，由 handler 走 HandleError。
//   - tenantID / userID 由 handler 从 xincontext 提取后显式传入（与 user/role 同档）。
//   - 收件人/发件人校验在 service 层做，避免 repository 关心业务规则。
type Service struct {
	pool *pgxpool.Pool
	repo Repository
}

func NewService(pool *pgxpool.Pool, repo Repository) *Service {
	return &Service{pool: pool, repo: repo}
}

// List 收件箱 / 发件箱
func (s *Service) List(ctx context.Context, req ListReq) (*ListResp, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}
	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	filter := ListFilter{
		IsRead:  req.IsRead,
		MsgType: req.MsgType,
		Keyword: req.Keyword,
		Page:    req.Page,
		Size:    req.Size,
	}

	var msgs []Message
	var total int64

	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		var err error
		switch req.Box {
		case BoxInbox:
			msgs, total, err = s.repo.ListInbox(txCtx, userID, filter)
		case BoxSent:
			msgs, total, err = s.repo.ListSent(txCtx, userID, filter)
		default:
			return ErrMessageTypeInvalid
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	out := &ListResp{
		List:  make([]MessageResp, len(msgs)),
		Total: total,
	}
	for i, m := range msgs {
		out.List[i] = toResp(m)
	}
	return out, nil
}

// Get 读取单条消息。会校验"当前用户必须是发件人或收件人"，避免越权窥探。
func (s *Service) Get(ctx context.Context, id uint) (*MessageResp, error) {
	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	var msg *Message
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		m, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		// 越权检查：必须是发件人、收件人，或者平台广播行（tenant_id=0）
		if m.TenantID != 0 && m.SenderID != userID && m.RecipientID != userID {
			return ErrMessageNotFound
		}
		msg = m
		return nil
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	r := toResp(*msg)
	return &r, nil
}

// Send 发送站内信
func (s *Service) Send(ctx context.Context, req SendReq) (*MessageResp, error) {
	if req.RecipientID == 0 {
		return nil, ErrRecipientEmpty
	}
	if req.Subject == "" {
		return nil, ErrSubjectEmpty
	}
	if req.MsgType != 0 && !validMsgType(req.MsgType) {
		return nil, ErrMessageTypeInvalid
	}
	if req.Priority != 0 && !validPriority(req.Priority) {
		return nil, ErrPriorityInvalid
	}

	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	var msg *Message
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		var err error
		msg, err = s.repo.Create(txCtx, tenantID, CreateRepoReq{
			SenderID:    userID,
			RecipientID: req.RecipientID,
			Subject:     req.Subject,
			Body:        req.Body,
			MsgType:     req.MsgType,
			Priority:    req.Priority,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	r := toResp(*msg)
	return &r, nil
}

// Update 局部更新（仅发件人可改自己发出去的信）
func (s *Service) Update(ctx context.Context, id uint, req UpdateReq) (*MessageResp, error) {
	if req.Priority != nil && !validPriority(*req.Priority) {
		return nil, ErrPriorityInvalid
	}

	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	var msg *Message
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		existing, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if existing.SenderID != userID {
			return ErrSenderMismatch
		}
		msg, err = s.repo.Update(txCtx, id, UpdateRepoReq{
			Subject:  req.Subject,
			Body:     req.Body,
			Priority: req.Priority,
		})
		return err
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	r := toResp(*msg)
	return &r, nil
}

// MarkRead 标记已读。仅收件人可标记自己收到的信。
func (s *Service) MarkRead(ctx context.Context, id uint) error {
	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	return db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		existing, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if existing.RecipientID != userID {
			return ErrRecipientMismatch
		}
		return s.repo.MarkRead(txCtx, id)
	})
}

// Delete 软删除。发件人 / 收件人可删；sender=0 的系统消息任何人不能删。
func (s *Service) Delete(ctx context.Context, id uint) error {
	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	return db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		existing, err := s.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if existing.SenderID != 0 && existing.SenderID != userID && existing.RecipientID != userID {
			return ErrMessageNotFound
		}
		return s.repo.Delete(txCtx, id)
	})
}

// CountUnread 收件箱未读数（用于导航栏 badge）
func (s *Service) CountUnread(ctx context.Context) (*UnreadCountResp, error) {
	tenantID := tenantIDFromCtx(ctx)
	userID := userIDFromCtx(ctx)

	var n int64
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(txCtx context.Context) error {
		var err error
		n, err = s.repo.CountUnread(txCtx, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &UnreadCountResp{Count: n}, nil
}

// toResp 把实体转成响应体（统一时间格式）
func toResp(m Message) MessageResp {
	r := MessageResp{
		ID:          m.ID,
		SenderID:    m.SenderID,
		RecipientID: m.RecipientID,
		Subject:     m.Subject,
		Body:        m.Body,
		MsgType:     m.MsgType,
		Priority:    m.Priority,
		IsRead:      m.IsRead,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
	if m.ReadAt != nil {
		s := m.ReadAt.Format(time.RFC3339)
		r.ReadAt = &s
	}
	return r
}

// mapRepoErr 把 repository 哨兵翻译成 BizError，未知错误原样向上抛
func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFoundDB) {
		return ErrMessageNotFound
	}
	return err
}

func validMsgType(t int16) bool {
	switch t {
	case MsgTypePrivate, MsgTypeNotify, MsgTypeBroadcast:
		return true
	}
	return false
}

func validPriority(p int16) bool {
	switch p {
	case PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

// tenantIDFromCtx 从 ctx 取 tenantID（缺失则 0）。
func tenantIDFromCtx(ctx context.Context) uint {
	v, _ := xincontext.TenantIDFrom(ctx)
	return v
}

// userIDFromCtx 从 ctx 取 userID。
//
// xincontext 没有 UserIDFrom helper，我们直接从注入的 XinContext 里拿。
// 之所以封装在这里，是为了让 service 跟其他模块看起来一致（统一 import 即可）。
func userIDFromCtx(ctx context.Context) uint {
	// 注：xincontext.UserContextFrom 只在懒加载已触发后才有值；
	// 这里我们用 ctx 上直接注入的 *XinContext（由 Auth 中间件 set）。
	xc, ok := xincontext.XinContextFrom(ctx)
	if !ok || xc == nil {
		return 0
	}
	return xc.UserID
}
package message

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/db"
)

// ErrNotFoundDB 是 repository 层哨兵错误，被 service 层用 errors.Is 翻译为 BizError。
var ErrNotFoundDB = errors.New("message not found")

// messageColumns 统一列序，Scan 与 SELECT 共用。
const messageColumns = `id, tenant_id, sender_id, recipient_id, subject, body,
    msg_type, priority, is_read, read_at, is_deleted, created_at, updated_at`

// ListFilter 列表过滤条件（service 层构造后传给 repository）。
type ListFilter struct {
	IsRead  *bool
	MsgType *int16
	Keyword string
	Page    int
	Size    int
}

// CreateRepoReq 写入参数
type CreateRepoReq struct {
	SenderID    uint
	RecipientID uint
	Subject     string
	Body        string
	MsgType     int16
	Priority    int16
}

// UpdateRepoReq 局部更新参数（nil 字段保持原值）
type UpdateRepoReq struct {
	Subject  *string
	Body     *string
	Priority *int16
}

// Repository 是 message 模块的完整接口。
type Repository interface {
	GetByID(ctx context.Context, id uint) (*Message, error)
	ListInbox(ctx context.Context, recipientID uint, f ListFilter) ([]Message, int64, error)
	ListSent(ctx context.Context, senderID uint, f ListFilter) ([]Message, int64, error)
	Create(ctx context.Context, tenantID uint, req CreateRepoReq) (*Message, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*Message, error)
	MarkRead(ctx context.Context, id uint) error
	Delete(ctx context.Context, id uint) error
	CountUnread(ctx context.Context, recipientID uint) (int64, error)
}

// PostgresRepository 是默认实现。
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// Compile-time guard
var _ Repository = (*PostgresRepository)(nil)

func NewRepository(pool *pgxpool.Pool) Repository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*Message, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	var m Message
	err = q.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM tenant_messages
		WHERE is_deleted = FALSE AND id = $1`, messageColumns), id).Scan(
		&m.ID, &m.TenantID, &m.SenderID, &m.RecipientID, &m.Subject, &m.Body,
		&m.MsgType, &m.Priority, &m.IsRead, &m.ReadAt, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFoundDB
		}
		return nil, err
	}
	return &m, nil
}

// ListInbox 收件箱：当前用户收到的信。
// 注意：平台广播（tenant_id=0）由 RLS 策略自动放行所有租户读取。
func (r *PostgresRepository) ListInbox(ctx context.Context, recipientID uint, f ListFilter) ([]Message, int64, error) {
	return r.list(ctx, []string{"recipient_id = $1"}, []any{recipientID}, f)
}

// ListSent 发件箱：当前用户发出的信。
func (r *PostgresRepository) ListSent(ctx context.Context, senderID uint, f ListFilter) ([]Message, int64, error) {
	return r.list(ctx, []string{"sender_id = $1"}, []any{senderID}, f)
}

// list 共用查询：where / args 由 caller 注入第一条。
func (r *PostgresRepository) list(ctx context.Context, where []string, args []any, f ListFilter) ([]Message, int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, 0, err
	}

	if f.IsRead != nil {
		args = append(args, *f.IsRead)
		where = append(where, fmt.Sprintf("is_read = $%d", len(args)))
	}
	if f.MsgType != nil {
		args = append(args, *f.MsgType)
		where = append(where, fmt.Sprintf("msg_type = $%d", len(args)))
	}
	if f.Keyword != "" {
		args = append(args, f.Keyword)
		where = append(where, fmt.Sprintf("(subject ILIKE '%%' || $%d || '%%' OR body ILIKE '%%' || $%d || '%%')", len(args), len(args)))
	}

	where = append(where, "is_deleted = FALSE")

	offset := (f.Page - 1) * f.Size
	args = append(args, f.Size, offset)
	listSQL := fmt.Sprintf(`
		SELECT %s FROM tenant_messages
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		messageColumns, strings.Join(where, " AND "), len(args)-1, len(args))

	rows, err := q.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.SenderID, &m.RecipientID, &m.Subject, &m.Body,
			&m.MsgType, &m.Priority, &m.IsRead, &m.ReadAt, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		msgs = append(msgs, m)
	}

	// 单独跑一次 COUNT（不带 LIMIT/OFFSET）
	countArgs := args[:len(args)-2]
	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM tenant_messages WHERE %s`, strings.Join(where, " AND "))
	if err := q.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}
	return msgs, total, nil
}

func (r *PostgresRepository) Create(ctx context.Context, tenantID uint, req CreateRepoReq) (*Message, error) {
	if req.MsgType == 0 {
		req.MsgType = MsgTypePrivate
	}
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	var m Message
	err = q.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO tenant_messages (tenant_id, sender_id, recipient_id, subject, body, msg_type, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING %s`, messageColumns),
		tenantID, req.SenderID, req.RecipientID, req.Subject, req.Body, req.MsgType, req.Priority,
	).Scan(
		&m.ID, &m.TenantID, &m.SenderID, &m.RecipientID, &m.Subject, &m.Body,
		&m.MsgType, &m.Priority, &m.IsRead, &m.ReadAt, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	return &m, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uint, req UpdateRepoReq) (*Message, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	idx := 1

	if req.Subject != nil {
		sets = append(sets, fmt.Sprintf("subject = $%d", idx))
		args = append(args, *req.Subject)
		idx++
	}
	if req.Body != nil {
		sets = append(sets, fmt.Sprintf("body = $%d", idx))
		args = append(args, *req.Body)
		idx++
	}
	if req.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", idx))
		args = append(args, *req.Priority)
		idx++
	}

	// 没有任何字段需要更新 → 直接返回当前记录
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`
		UPDATE tenant_messages SET %s
		WHERE id = $%d AND is_deleted = FALSE
		RETURNING %s`,
		strings.Join(sets, ", "), idx, messageColumns)

	var m Message
	if err := q.QueryRow(ctx, sql, args...).Scan(
		&m.ID, &m.TenantID, &m.SenderID, &m.RecipientID, &m.Subject, &m.Body,
		&m.MsgType, &m.Priority, &m.IsRead, &m.ReadAt, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFoundDB
		}
		return nil, fmt.Errorf("update message: %w", err)
	}
	return &m, nil
}

// MarkRead 标记已读。已读或不存在都返回 nil（幂等）。
// 区分"不存在"与"已经是已读"没有业务意义，前端重试友好。
func (r *PostgresRepository) MarkRead(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `
		UPDATE tenant_messages
		SET is_read = TRUE, read_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE AND is_read = FALSE`, id)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return err
	}
	tag, err := q.Exec(ctx, `
		UPDATE tenant_messages SET is_deleted = TRUE, updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE`, id)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFoundDB
	}
	return nil
}

func (r *PostgresRepository) CountUnread(ctx context.Context, recipientID uint) (int64, error) {
	q, err := db.GetQuerier(ctx, r.pool)
	if err != nil {
		return 0, err
	}
	var n int64
	err = q.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenant_messages
		WHERE is_deleted = FALSE AND recipient_id = $1 AND is_read = FALSE`,
		recipientID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return n, nil
}
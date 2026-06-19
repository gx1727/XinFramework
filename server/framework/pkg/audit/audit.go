// Package audit 提供统一的 db_logs 写入入口。
//
// 设计目标：
//   - 一行 Log 调用记录一条审计
//   - 失败不抛：业务路径不应被审计写库失败打断
//   - 自动从 XinContext 提取 TenantID / UserID
//   - OldData / NewData 字段是 any，内部用 JSON 序列化
package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	xincontext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/logger"
)

// Entry 一条审计记录。TenantID / UserID 留 0 时从 XinContext 取。
type Entry struct {
	TenantID  uint
	UserID    uint
	Action    string // 例: "user:org_change" / "org:delete"
	TableName string // 例: "users" / "organizations"
	RecordID  uint   // 被操作记录的主键
	OldData   any    // 改动前的快照（任意 JSON 可序列化的结构）
	NewData   any    // 改动后的快照
	IP        string // 留空时从 ctx 取
}

// Log 写一条审计记录。错误仅记日志，不返回给业务路径。
//
// Phase 4: 显式传入 pool，不再依赖 db.GetQuerier(ctx) 全局访问器。
func Log(ctx context.Context, pool *pgxpool.Pool, e Entry) {
	tenantID := e.TenantID
	userID := e.UserID
	if xc, ok := xincontext.XinContextFrom(ctx); ok && xc != nil {
		if tenantID == 0 {
			tenantID = xc.TenantID
		}
		if userID == 0 {
			userID = xc.UserID
		}
	}
	ip := e.IP
	if ip == "" {
		ip = IPFrom(ctx)
	}

	oldJSON, _ := json.Marshal(e.OldData)
	newJSON, _ := json.Marshal(e.NewData)

	querier, err := db.GetQuerier(ctx, pool)
	if err != nil {
		logger.Module("audit").Warnf("[Log] no querier: %v", err)
		return
	}

	// db_logs 没有 is_deleted 列，没有 created_by 列，但有 tenant_id / user_id。
	// RLS 在 db_logs 上没启用（参见 001_framework_init.sql），所以这里不强制租户上下文。
	// user_id 允许 NULL，IP 允许 NULL。
	// old_data / new_data 是 JSONB；pgx 把 string 当 text 发，必须 ::jsonb 显式 cast。
	_, err = querier.Exec(ctx, `
		INSERT INTO db_logs (tenant_id, user_id, action, table_name, record_id, old_data, new_data, ip, created_at)
		VALUES ($1, NULLIF($2, 0), $3, $4, NULLIF($5, 0), NULLIF($6, '')::jsonb, NULLIF($7, '')::jsonb, NULLIF($8, ''), NOW())`,
		tenantID,
		userID,
		e.Action,
		e.TableName,
		e.RecordID,
		string(oldJSON),
		string(newJSON),
		ip,
	)
	if err != nil {
		logger.Module("audit").Warnf("[Log] insert failed: %v | action=%s table=%s record=%d", err, e.Action, e.TableName, e.RecordID)
	}
}

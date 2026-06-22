// Package config 通用配置 - 服务层
//
// 与 apps/reference/dict 的 service 架构对标：
//
//	| 操作          | 事务              | 缓存策略                  |
//	| CreateGroup   | RunInPlatformTx    | InvalidateAll             |
//	| UpdateGroup   | RunInPlatformTx    | InvalidateAll             |
//	| DeleteGroup   | RunInPlatformTx    | InvalidateAll             |
//	| CreateItem    | RunInPlatformTx    | InvalidateAll             |
//	| UpdateItem    | RunInPlatformTx    | InvalidateAll             |
//	| DeleteItem    | RunInPlatformTx    | InvalidateAll             |
//	| Resolve       | RunInTenantTx      | 命中即返 / miss 则加载    |
//	| UpsertOverride| RunInTenantTx      | Invalidate(tenantID)      |
//	| DeleteOverride| RunInTenantTx      | Invalidate(tenantID)      |
//	| UpsertVisibility | RunInPlatformTx | Invalidate(tenantID)      |
//
// 所有写操作走 RunInPlatformTx（platform 域，bypass RLS）。
// Resolve 走 RunInTenantTx（受 RLS 约束）。
package config

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/resp"
)

// Service 是 config 模块的核心 service。
//
// 持有 pool + repo + cache 三件套。
// 同时承担 platform 和 tenant 两域的操作——
// 区分靠事务上下文（RunInPlatformTx vs RunInTenantTx）。
type Service struct {
	pool  *pgxpool.Pool
	repo  ConfigRepository
	cache *Cache
}

func NewService(pool *pgxpool.Pool, repo ConfigRepository, cache *Cache) *Service {
	return &Service{pool: pool, repo: repo, cache: cache}
}

// ============================================================================
// Group — Platform 域
// ============================================================================

// CreateGroup 创建配置分组（super_admin 调用）
func (s *Service) CreateGroup(ctx context.Context, req createGroupRequest, scope string) (*ConfigCategory, error) {
	if !isValidScope(scope) {
		return nil, ErrInvalidVisibility
	}
	var out *ConfigCategory
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		repoReq := CreateGroupRepoReq{
			Code:        req.Code,
			Name:        req.Name,
			Description: req.Description,
			Icon:        req.Icon,
			Sort:        req.Sort,
			IsSystem:    req.IsSystem,
			IsPublic:    req.IsPublic,
		}
		tenantID := uint(0)
		if scope == "tenant" {
			tenantID = req.TenantID
		}
		var err error
		out, err = s.repo.CreateGroup(ctx, tenantID, scope, repoReq)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_category:create", TableName: "config_categories", RecordID: uint(out.ID), OldData: nil, NewData: out})
	return out, nil
}

// UpdateGroup 修改配置分组（super_admin 调用）
func (s *Service) UpdateGroup(ctx context.Context, id uint, req updateGroupRequest) (*ConfigCategory, error) {
	var out *ConfigCategory
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		out, err = s.repo.UpdateGroup(ctx, id, UpdateGroupRepoReq{
			Name:        req.Name,
			Description: req.Description,
			Icon:        req.Icon,
			Sort:        req.Sort,
			IsPublic:    req.IsPublic,
			Visibility:  req.Visibility,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_category:update", TableName: "config_categories", RecordID: uint(id), OldData: nil, NewData: out})
	return out, nil
}

// DeleteGroup 删除配置分组（super_admin 调用）
func (s *Service) DeleteGroup(ctx context.Context, id uint) error {
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		// 前置校验：group 下是否有未删除的 item
		n, err := s.repo.CountItemsByGroup(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w (item数=%d)", ErrGroupHasItems, n)
		}
		return s.repo.DeleteGroup(ctx, id)
	})
	if err != nil {
		return mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_category:delete", TableName: "config_categories", RecordID: uint(id), OldData: nil, NewData: nil})
	return nil
}

// ListPlatformGroups 列全部平台 group
func (s *Service) ListPlatformGroups(ctx context.Context) ([]ConfigCategory, error) {
	var out []ConfigCategory
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		out, err = s.repo.ListPlatformGroups(ctx)
		return err
	})
	return out, err
}

// ============================================================================
// Item — Platform 域
// ============================================================================

// CreateItem 创建平台 item
func (s *Service) CreateItem(ctx context.Context, categoryID uint, req createItemRequest) (*ConfigItem, error) {
	if !isValidType(req.Type) {
		return nil, ErrInvalidItemType
	}
	if err := validateValueForType(req.Value, req.Type, req.Options); err != nil {
		return nil, err
	}
	var out *ConfigItem
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		repoReq := CreateItemRepoReq{
			Key:          req.Key,
			Value:        req.Value,
			DefaultValue: req.DefaultValue,
			Type:         req.Type,
			Label:        req.Label,
			Description:  req.Description,
			Options:      req.Options,
			Validation:   req.Validation,
			Sort:         req.Sort,
			IsPublic:     req.IsPublic,
			IsReadonly:   req.IsReadonly,
			IsSystem:     req.IsSystem,
		}
		var err error
		out, err = s.repo.CreateItem(ctx, 0, categoryID, repoReq)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_item:create", TableName: "config_items", RecordID: uint(out.ID), OldData: nil, NewData: out})
	return out, nil
}

// UpdateItem 修改平台 item
func (s *Service) UpdateItem(ctx context.Context, id uint, req updateItemRequest) (*ConfigItem, error) {
	var before, after *ConfigItem
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		before, err = s.repo.GetItemByID(ctx, id)
		if err != nil {
			return err
		}
		after, err = s.repo.UpdateItem(ctx, id, UpdateItemRepoReq{
			Value:       req.Value,
			Label:       req.Label,
			Description: req.Description,
			Sort:        req.Sort,
			IsPublic:    req.IsPublic,
			IsReadonly:  req.IsReadonly,
			Status:      req.Status,
		})
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_item:update", TableName: "config_items", RecordID: uint(id), OldData: before, NewData: after})
	return after, nil
}

// DeleteItem 删除平台 item
func (s *Service) DeleteItem(ctx context.Context, id uint) error {
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		// 前置校验：是否有租户覆盖此 platform item
		// （暂简化：直接删，由 SQL 索引 uk_config_item_override 兜底）
		return s.repo.DeleteItem(ctx, id)
	})
	if err != nil {
		return mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.InvalidateAll()
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_item:delete", TableName: "config_items", RecordID: uint(id), OldData: nil, NewData: nil})
	return nil
}

// ListPlatformItems 列某平台 group 的所有 item
func (s *Service) ListPlatformItems(ctx context.Context, categoryID uint) ([]ConfigItem, error) {
	var out []ConfigItem
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		out, err = s.repo.ListPlatformItemsByGroup(ctx, categoryID)
		return err
	})
	return out, err
}

// ============================================================================
// Override — Tenant 域
// ============================================================================

// UpsertOverride 租户 upsert 对某 platform item 的 value 覆盖
func (s *Service) UpsertOverride(ctx context.Context, tenantID, platformItemID uint, value interface{}) (*ConfigItem, error) {
	var out *ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		// 前置校验：platform item 必须存在
		var platformItem ConfigItem
		err := s.pool.QueryRow(ctx, `
			SELECT id, is_readonly, is_system
			FROM config_items
			WHERE id = $1 AND tenant_id = 0 AND is_deleted = FALSE`, platformItemID).Scan(
			&platformItem.ID, &platformItem.IsReadonly, &platformItem.IsSystem)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlatformItemMismatch
			}
			return err
		}
		if platformItem.IsReadonly || platformItem.IsSystem {
			return ErrGroupReadonly
		}

		var innerErr error
		out, innerErr = s.repo.UpsertOverride(ctx, tenantID, platformItemID, value)
		return innerErr
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.Invalidate(tenantID)
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_item:override_upsert", TableName: "config_items", RecordID: uint(out.ID), OldData: nil, NewData: out})
	return out, nil
}

// DeleteOverride 租户删除对 platform item 的覆盖
func (s *Service) DeleteOverride(ctx context.Context, tenantID, platformItemID uint) error {
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		return s.repo.DeleteOverride(ctx, tenantID, platformItemID)
	})
	if err != nil {
		return mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.Invalidate(tenantID)
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_item:override_delete", TableName: "config_items", RecordID: uint(platformItemID), OldData: nil, NewData: nil})
	return nil
}

// ============================================================================
// Visibility — Platform 域
// ============================================================================

// ListVisibility 列某 platform group 对各租户的访问级别
func (s *Service) ListVisibility(ctx context.Context, categoryID uint) ([]ConfigVisibility, error) {
	var out []ConfigVisibility
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		out, err = s.repo.ListVisibility(ctx, categoryID)
		return err
	})
	return out, err
}

// UpsertVisibility super_admin 设置某 platform group 对某租户的访问级别
func (s *Service) UpsertVisibility(ctx context.Context, categoryID, tenantID uint, access string) (*ConfigVisibility, error) {
	if !isValidAccess(access) {
		return nil, ErrInvalidAccess
	}
	var out *ConfigVisibility
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		var err error
		out, err = s.repo.UpsertVisibility(ctx, categoryID, tenantID, access)
		return err
	})
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Invalidate(tenantID)
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_category:visibility_upsert", TableName: "config_visibility", RecordID: uint(out.ID), OldData: nil, NewData: out})
	return out, nil
}

// DeleteVisibility 删除某 platform group 对某租户的访问级别（恢复默认 visibility 策略）
func (s *Service) DeleteVisibility(ctx context.Context, categoryID, tenantID uint) error {
	err := db.RunInPlatformTx(ctx, s.pool, func(ctx context.Context) error {
		return s.repo.DeleteVisibility(ctx, categoryID, tenantID)
	})
	if err != nil {
		return mapRepoError(err)
	}
	if s.cache != nil {
		s.cache.Invalidate(tenantID)
	}
	audit.Log(ctx, s.pool, audit.Entry{Action: "config_category:visibility_delete", TableName: "config_visibility", RecordID: uint(categoryID), OldData: nil, NewData: nil})
	return nil
}

// ============================================================================
// Resolve — Tenant 域（业务合并消费）
// ============================================================================

// Resolve 单条：按 group_code 取租户视角的合并配置
func (s *Service) Resolve(ctx context.Context, tenantID uint, groupCode string) (*ResolvedConfig, error) {
	if tenantID == 0 {
		return nil, ErrGroupInvisible
	}
	if groupCode == "" {
		return nil, fmt.Errorf("group code required")
	}

	// 命中缓存
	if s.cache != nil {
		if rc, ok := s.cache.Get(tenantID, groupCode); ok {
			return rc, nil
		}
	}

	var out *ResolvedConfig
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		out, err = s.repo.ResolveGroupForTenant(ctx, tenantID, groupCode)
		return err
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	// 写缓存（仅 platform/tenant 合并成功的）
	if s.cache != nil && out != nil {
		// 懒加载模式：当前只缓存这一个 group，不全量加载
		groups := map[string]*ResolvedConfig{groupCode: out}
		s.cache.Put(tenantID, groups)
	}
	return out, nil
}

// ResolveAll 取某租户的全部合并配置
func (s *Service) ResolveAll(ctx context.Context, tenantID uint) (map[string]*ResolvedConfig, error) {
	if tenantID == 0 {
		return nil, ErrGroupInvisible
	}

	// 懒加载
	groups, err := s.cache.LoadOrLoadAll(tenantID, func() (map[string]*ResolvedConfig, error) {
		var out map[string]*ResolvedConfig
		err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
			var err error
			out, err = s.repo.ResolveAllForTenant(ctx, tenantID)
			return err
		})
		if err != nil {
			return nil, mapRepoError(err)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// ============================================================================
// Public Read — Tenant 域（无需鉴权，公共读）
// ============================================================================

// ListPublicItems 列某租户的公开配置项（is_public=TRUE 的 group 下所有 item）
func (s *Service) ListPublicItems(ctx context.Context, tenantID uint) ([]ConfigItem, error) {
	var out []ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		out, err = s.repo.ListPublicItemsByTenant(ctx, tenantID)
		return err
	})
	return out, err
}

// ListPublicItemsByGroupCode 按 group code 取公开配置项
func (s *Service) ListPublicItemsByGroupCode(ctx context.Context, tenantID uint, groupCode string) ([]ConfigItem, error) {
	var out []ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		out, err = s.repo.ListPublicItemsByGroupCode(ctx, tenantID, groupCode)
		return err
	})
	return out, err
}

// ============================================================================
// helpers
// ============================================================================

// mapRepoError 把 repo 层 pg 错误映射到业务 resp.Err。
// 当前大部分错误已经由 repository 直接返回 resp.Err（如 ErrGroupNotFound 等），
// 这里只兜底 SQL unique_violation / FK 约束等。
func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	// 已经是 *resp.BizError 直接返回
	var bizErr *resp.BizError
	if errors.As(err, &bizErr) {
		return err
	}
	// PG unique violation 23505 → code exists
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			msg := pgErr.Message
			if strings.Contains(msg, "uk_config_group_code") {
				return ErrGroupCodeExists
			}
			if strings.Contains(msg, "uk_config_item_key") {
				return ErrItemKeyExists
			}
			if strings.Contains(msg, "uk_config_visibility") {
				return ErrInvalidVisibility
			}
		}
	}
	return err
}

// isValidScope / isValidType / isValidAccess / isValidVisibility
// 取值校验，与 dict 对齐（defensive validation）。
func isValidScope(s string) bool      { return s == "platform" || s == "tenant" }
func isValidType(t string) bool       { return t == "string" || t == "number" || t == "boolean" || t == "json" || t == "select" || t == "multiselect" || t == "color" || t == "image" || t == "text" || t == "password" }
func isValidAccess(a string) bool     { return a == "invisible" || a == "readonly" || a == "editable" }
func isValidVisibility(v string) bool { return v == "all" || v == "whitelist" || v == "blacklist" }

// validateValueForType 校验 value 与 type 匹配（与 dict 的 validate 一致）
func validateValueForType(value interface{}, typ string, options interface{}) error {
	if value == nil {
		return nil
	}
	switch typ {
	case "string", "text", "password", "color", "image":
		if _, ok := value.(string); !ok {
			return ErrInvalidValueForType
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// ok
		default:
			return ErrInvalidValueForType
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return ErrInvalidValueForType
		}
	case "json":
		// 任何值都可
	case "select", "multiselect":
		if options == nil {
			return nil
		}
		// 简化：不严格校验 options 范围
	}
	return nil
}

// 时间戳默认值（用于 audit）—— 防止 import 抖动
var _ = time.Now

// satisfy imports —— 防止 unused import 报错
var (
	_ = audit.Log
	_ = pgx.ErrNoRows
)

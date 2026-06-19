// Package config 通用配置 - 服务层
package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/db"
)

type Service struct {
	pool  *pgxpool.Pool
	repo  ConfigRepository
	cache *Cache
}

func NewService(pool *pgxpool.Pool, repo ConfigRepository, cache *Cache) *Service {
	return &Service{pool: pool, repo: repo, cache: cache}
}

// =============== Group ===============

func (s *Service) ListGroups(ctx context.Context, tenantID uint) ([]ConfigGroup, error) {
	var groups []ConfigGroup
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		groups, err = s.repo.ListGroups(ctx, tenantID)
		return err
	})
	return groups, err
}

func (s *Service) CreateGroup(ctx context.Context, tenantID uint, req createGroupRequest) (*ConfigGroup, error) {
	var g *ConfigGroup
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		created, err := s.repo.CreateGroup(ctx, tenantID, CreateGroupRepoReq{
			Code:        req.Code,
			Name:        req.Name,
			Description: req.Description,
			Icon:        req.Icon,
			Sort:        req.Sort,
			IsSystem:    false, // 业务创建的永远不是系统
			IsPublic:    req.IsPublic,
		})
		if err != nil {
			return err
		}
		g = created
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_group:create",
			TableName: "config_groups",
			RecordID:  g.ID,
			NewData: map[string]any{
				"code": g.Code, "name": g.Name, "is_public": g.IsPublic,
			},
		})
		return nil
	})
	return g, err
}

func (s *Service) UpdateGroup(ctx context.Context, tenantID, id uint, req updateGroupRequest) (*ConfigGroup, error) {
	var g *ConfigGroup
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetGroupByID(ctx, id)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrGroupNotFound
		}
		updated, err := s.repo.UpdateGroup(ctx, id, UpdateGroupRepoReq{
			Name:        req.Name,
			Description: req.Description,
			Icon:        req.Icon,
			Sort:        req.Sort,
			IsPublic:    req.IsPublic,
			Status:      req.Status,
		})
		if err != nil {
			return err
		}
		g = updated
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_group:update",
			TableName: "config_groups",
			RecordID:  g.ID,
			OldData:   map[string]any{"name": old.Name, "sort": old.Sort, "is_public": old.IsPublic},
			NewData:   map[string]any{"name": g.Name, "sort": g.Sort, "is_public": g.IsPublic},
		})
		return nil
	})
	if err == nil {
		s.cache.Invalidate(tenantID)
	}
	return g, err
}

func (s *Service) DeleteGroup(ctx context.Context, tenantID, id uint) error {
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		g, err := s.repo.GetGroupByID(ctx, id)
		if err != nil {
			return err
		}
		if g.TenantID != tenantID {
			return ErrGroupNotFound
		}
		if g.IsSystem {
			return ErrGroupIsSystem
		}
		// 保护：组下还有项
		n, err := s.repo.CountItemsByGroup(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("%w (item count=%d)", ErrGroupHasItems, n)
		}
		if err := s.repo.DeleteGroup(ctx, id); err != nil {
			return err
		}
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_group:delete",
			TableName: "config_groups",
			RecordID:  g.ID,
			OldData:   map[string]any{"code": g.Code, "name": g.Name},
		})
		return nil
	})
	if err == nil {
		s.cache.Invalidate(tenantID)
	}
	return err
}

// =============== Item ===============

func (s *Service) ListItemsByGroup(ctx context.Context, tenantID, groupID uint) ([]ConfigItem, error) {
	var items []ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		g, err := s.repo.GetGroupByID(ctx, groupID)
		if err != nil {
			return err
		}
		if g.TenantID != 0 && g.TenantID != tenantID {
			return ErrGroupNotFound
		}
		items, err = s.repo.ListItemsByGroup(ctx, groupID)
		return err
	})
	return items, err
}

func (s *Service) ListItemsByTenant(ctx context.Context, tenantID uint) ([]ConfigItem, error) {
	var items []ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		items, err = s.repo.ListItemsByTenant(ctx, tenantID)
		return err
	})
	return items, err
}

func (s *Service) CreateItem(ctx context.Context, tenantID, groupID uint, req createItemRequest) (*ConfigItem, error) {
	// 业务层校验
	if err := validateValueForType(req.Type, req.Value, req.Options); err != nil {
		return nil, err
	}
	var item *ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		g, err := s.repo.GetGroupByID(ctx, groupID)
		if err != nil {
			return err
		}
		if g.TenantID != 0 && g.TenantID != tenantID {
			return ErrGroupNotFound
		}
		// value 默认等于 default_value
		value := req.Value
		if value == nil {
			value = req.DefaultValue
		}
		created, err := s.repo.CreateItem(ctx, tenantID, groupID, CreateItemRepoReq{
			Key:          req.Key,
			Value:        value,
			DefaultValue: req.DefaultValue,
			Type:         req.Type,
			Label:        req.Label,
			Description:  req.Description,
			Options:      req.Options,
			Validation:   req.Validation,
			Sort:         req.Sort,
			IsPublic:     req.IsPublic,
			IsReadonly:   req.IsReadonly,
			IsSystem:     false,
		})
		if err != nil {
			return err
		}
		item = created
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_item:create",
			TableName: "config_items",
			RecordID:  item.ID,
			NewData: map[string]any{
				"key": item.Key, "type": item.Type, "group_id": item.GroupID,
			},
		})
		return nil
	})
	if err == nil {
		s.cache.Invalidate(tenantID)
	}
	return item, err
}

func (s *Service) UpdateItem(ctx context.Context, tenantID, id uint, req updateItemRequest) (*ConfigItem, error) {
	var item *ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetItemByID(ctx, id)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrItemNotFound
		}
		if old.IsReadonly {
			return ErrItemIsReadonly
		}
		// 业务层校验 value
		if req.Value != nil {
			if err := validateValueForType(old.Type, *req.Value, old.Options); err != nil {
				return err
			}
		}
		updated, err := s.repo.UpdateItem(ctx, id, UpdateItemRepoReq{
			Value:       req.Value,
			Label:       req.Label,
			Description: req.Description,
			Sort:        req.Sort,
			IsPublic:    req.IsPublic,
			IsReadonly:  req.IsReadonly,
			Status:      req.Status,
		})
		if err != nil {
			return err
		}
		item = updated
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_item:update",
			TableName: "config_items",
			RecordID:  item.ID,
			OldData:   map[string]any{"value": old.Value, "sort": old.Sort, "is_public": old.IsPublic},
			NewData:   map[string]any{"value": item.Value, "sort": item.Sort, "is_public": item.IsPublic},
		})
		return nil
	})
	if err == nil {
		s.cache.Invalidate(tenantID)
	}
	return item, err
}

func (s *Service) ResetItem(ctx context.Context, tenantID, id uint) (*ConfigItem, error) {
	var item *ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetItemByID(ctx, id)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrItemNotFound
		}
		if old.IsReadonly {
			return ErrItemIsReadonly
		}
		updated, err := s.repo.ResetItem(ctx, id)
		if err != nil {
			return err
		}
		item = updated
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_item:reset",
			TableName: "config_items",
			RecordID:  item.ID,
			OldData:   map[string]any{"value": old.Value},
			NewData:   map[string]any{"value": item.Value},
		})
		return nil
	})
	if err == nil {
		s.cache.Invalidate(tenantID)
	}
	return item, err
}

func (s *Service) DeleteItem(ctx context.Context, tenantID, id uint) error {
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetItemByID(ctx, id)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrItemNotFound
		}
		if old.IsSystem {
			return ErrItemIsSystem
		}
		if err := s.repo.DeleteItem(ctx, id); err != nil {
			return err
		}
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "config_item:delete",
			TableName: "config_items",
			RecordID:  old.ID,
			OldData:   map[string]any{"key": old.Key, "group_id": old.GroupID},
		})
		return nil
	})
	if err == nil {
		s.cache.Invalidate(tenantID)
	}
	return err
}

// =============== Public 公共读（无租户上下文，可走任意 tenantID）===============

// GetPublicByGroup 公共读：未登录时调用，按 groupCode 取所有 is_public 项
func (s *Service) GetPublicByGroup(ctx context.Context, tenantID uint, groupCode string) (map[string]interface{}, error) {
	type cacheKey struct {
		tenantID  uint
		groupCode string
	}
	_ = cacheKey{} // 占位避免 unused
	// 优先走缓存
	if s.cache != nil {
		if all, ok := s.cache.GetAll(tenantID); ok {
			if items, ok := all[groupCode]; ok {
				return flattenItems(items), nil
			}
		}
	}
	// 缓存未命中：拉全量 + 缓存（这里仅 public 项）
	var publicAll []ConfigItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		publicAll, err = s.repo.ListPublicItemsByTenant(ctx, tenantID)
		return err
	})
	if err != nil {
		return nil, err
	}
	// 写缓存
	if s.cache != nil {
		groups := map[string][]*ConfigItem{}
		for i := range publicAll {
			it := publicAll[i]
			g, err := s.repo.GetGroupByID(ctx, it.GroupID)
			if err != nil {
				continue
			}
			groups[g.Code] = append(groups[g.Code], &it)
		}
		s.cache.PutAll(tenantID, groups)
		// 再取值
		if items, ok := groups[groupCode]; ok {
			return flattenItems(items), nil
		}
	}
	// 缓存未启用或未命中
	out := map[string]interface{}{}
	for _, it := range publicAll {
		g, err := s.repo.GetGroupByID(ctx, it.GroupID)
		if err != nil {
			continue
		}
		if g.Code == groupCode {
			out[it.Key] = it.Value
		}
	}
	return out, nil
}

// flattenItems 把 items 列表压扁为 key→value
func flattenItems(items []*ConfigItem) map[string]interface{} {
	out := make(map[string]interface{}, len(items))
	for _, it := range items {
		out[it.Key] = it.Value
	}
	return out
}

// =============== 校验 ===============

// validateValueForType 根据 item.type 校验 value 是否合法
func validateValueForType(itemType string, value interface{}, options interface{}) error {
	if value == nil {
		return nil
	}
	switch itemType {
	case "string", "text", "password", "image", "color":
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: type=%s expects string", ErrInvalidValueForType, itemType)
		}
	case "number":
		switch v := value.(type) {
		case float64, int, int32, int64, json.Number:
			// OK
		default:
			_ = v
			return fmt.Errorf("%w: type=number expects numeric", ErrInvalidValueForType)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%w: type=boolean expects bool", ErrInvalidValueForType)
		}
	case "json":
		// 任意 JSON 都接受
	case "select":
		// 必须在 options 内
		if options == nil {
			return nil
		}
		if !valueInOptions(value, options) {
			return ErrValueNotInOptions
		}
	case "multiselect":
		if options == nil {
			return nil
		}
		if !valueInOptions(value, options) {
			return ErrValueNotInOptions
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidItemType, itemType)
	}
	return nil
}

func valueInOptions(value interface{}, options interface{}) bool {
	optsBytes, err := json.Marshal(options)
	if err != nil {
		return false
	}
	var opts []map[string]interface{}
	if err := json.Unmarshal(optsBytes, &opts); err != nil {
		return false
	}
	// single value
	if v, ok := value.(string); ok {
		for _, o := range opts {
			if fmt.Sprintf("%v", o["value"]) == v {
				return true
			}
		}
		return false
	}
	// multi
	if arr, ok := value.([]interface{}); ok {
		for _, a := range arr {
			found := false
			for _, o := range opts {
				if fmt.Sprintf("%v", o["value"]) == fmt.Sprintf("%v", a) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

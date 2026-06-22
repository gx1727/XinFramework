// Package dict 数据字典服务层
package dict

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/db"
	dictpkg "gx1727.com/xin/framework/pkg/dict"
)

type Service struct {
	repo DictRepository
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool, repo DictRepository) *Service {
	return &Service{repo: repo, pool: pool}
}

// ========== 字典主表 ==========

// List 字典列表（按租户过滤 + 关键字模糊匹配 code/name）
func (s *Service) List(ctx context.Context, tenantID uint, req listRequest) ([]Dict, int64, error) {
	var list []Dict
	var total int64
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		var err error
		list, total, err = s.repo.List(ctx, tenantID, trimKeyword(req.Keyword), req.Page, req.Size)
		return err
	})
	return list, total, err
}

// trimKeyword 修剪前后空白；JS 序列化的 "undefined"/"null" 视作空
func trimKeyword(s string) string {
	s2 := s
	if len(s2) > 0 && (s2[0] == ' ' || s2[len(s2)-1] == ' ') {
		// 简单 trim
		start, end := 0, len(s2)
		for start < end && s2[start] == ' ' {
			start++
		}
		for end > start && s2[end-1] == ' ' {
			end--
		}
		s2 = s2[start:end]
	}
	if s2 == "undefined" || s2 == "null" {
		return ""
	}
	return s2
}

// Get 获取单个字典（跨租户校验）
func (s *Service) Get(ctx context.Context, tenantID uint, id uint) (*Dict, error) {
	var d *Dict
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		got, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if got.TenantID != 0 && got.TenantID != tenantID {
			return ErrDictNotFound
		}
		d = got
		return nil
	})
	return d, err
}

// Create 新建字典
func (s *Service) Create(ctx context.Context, tenantID uint, req createRequest) (*Dict, error) {
	var d *Dict
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		got, err := s.repo.Create(ctx, tenantID, CreateDictRepoReq{
			Code:   req.Code,
			Name:   req.Name,
			Sort:   req.Sort,
			Status: 1,
			Extend: req.Extend,
		})
		if err != nil {
			return err
		}
		d = got
		return nil
	})
	return d, err
}

// Update 更新字典
func (s *Service) Update(ctx context.Context, tenantID, id uint, req updateRequest) (*Dict, error) {
	var d *Dict
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrDictNotFound
		}

		status := req.Status
		if status == 0 {
			status = 1
		}

		updated, err := s.repo.Update(ctx, id, UpdateDictRepoReq{
			Name:   req.Name,
			Sort:   req.Sort,
			Status: status,
			Extend: req.Extend,
		})
		if err != nil {
			return err
		}
		d = updated
		return nil
	})
	return d, err
}

// Delete 删除字典前先检查字典项是否还有未删的；通过后写审计。
func (s *Service) Delete(ctx context.Context, tenantID, id uint) error {
	return db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		d, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if d.TenantID != tenantID {
			return ErrDictNotFound
		}

		n, err := s.repo.CountItems(ctx, id)
		if err != nil {
			return fmt.Errorf("count items: %w", err)
		}
		if n > 0 {
			return fmt.Errorf("%w (字典项数=%d)", ErrDictHasItems, n)
		}

		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}

		// 审计：删除字典
		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "dict:delete",
			TableName: "dicts",
			RecordID:  d.ID,
			OldData: map[string]any{
				"id":     d.ID,
				"code":   d.Code,
				"name":   d.Name,
				"status": d.Status,
			},
		})
		return nil
	})
}

// ========== 字典项 ==========

// ListItems 列出某字典下的所有字典项
func (s *Service) ListItems(ctx context.Context, tenantID, dictID uint) ([]DictItem, error) {
	var items []DictItem
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		d, err := s.repo.GetByID(ctx, dictID)
		if err != nil {
			return err
		}
		if d.TenantID != 0 && d.TenantID != tenantID {
			return ErrDictNotFound
		}
		items, err = s.repo.ListItems(ctx, dictID)
		return err
	})
	return items, err
}

// CreateItem 在字典下新增字典项；成功后刷缓存
func (s *Service) CreateItem(ctx context.Context, tenantID, dictID uint, req createItemRequest) (*DictItem, error) {
	var item *DictItem
	var dictCode string
	err := db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		d, err := s.repo.GetByID(ctx, dictID)
		if err != nil {
			return err
		}
		if d.TenantID != tenantID {
			return ErrDictNotFound
		}

		it, err := s.repo.CreateItem(ctx, tenantID, dictID, CreateDictItemRepoReq{
			Code:   req.Code,
			Name:   req.Name,
			Sort:   req.Sort,
			Status: 1,
			Extend: req.Extend,
		})
		if err != nil {
			return err
		}
		item = it
		dictCode = d.Code
		return nil
	})
	if err != nil {
		return nil, err
	}

	refreshDictCache(ctx, tenantID, dictCode)
	return item, nil
}

// UpdateItem 更新字典项；写审计 + 刷缓存
func (s *Service) UpdateItem(ctx context.Context, tenantID, itemID uint, req updateItemRequest) error {
	return db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetItemByID(ctx, itemID)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrDictItemNotFound
		}

		d, err := s.repo.GetByID(ctx, old.DictID)
		if err != nil {
			return err
		}

		status := req.Status
		if status == 0 {
			status = 1
		}

		if err := s.repo.UpdateItem(ctx, itemID, UpdateDictItemRepoReq{
			Name:   req.Name,
			Sort:   req.Sort,
			Status: status,
			Extend: req.Extend,
		}); err != nil {
			return err
		}

		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "dict_item:update",
			TableName: "dict_items",
			RecordID:  old.ID,
			OldData: map[string]any{
				"id":     old.ID,
				"code":   old.Code,
				"name":   old.Name,
				"sort":   old.Sort,
				"status": old.Status,
			},
			NewData: map[string]any{
				"name":   req.Name,
				"sort":   req.Sort,
				"status": status,
			},
		})

		refreshDictCache(ctx, tenantID, d.Code)
		return nil
	})
}

// DeleteItem 软删字典项；写审计 + 刷缓存
func (s *Service) DeleteItem(ctx context.Context, tenantID, itemID uint) error {
	return db.RunInTenantTx(ctx, s.pool, tenantID, func(ctx context.Context) error {
		old, err := s.repo.GetItemByID(ctx, itemID)
		if err != nil {
			return err
		}
		if old.TenantID != tenantID {
			return ErrDictItemNotFound
		}

		d, err := s.repo.GetByID(ctx, old.DictID)
		if err != nil {
			return err
		}

		if err := s.repo.DeleteItem(ctx, itemID); err != nil {
			return err
		}

		audit.Log(ctx, s.pool, audit.Entry{
			Action:    "dict_item:delete",
			TableName: "dict_items",
			RecordID:  old.ID,
			OldData: map[string]any{
				"id":      old.ID,
				"dict_id": old.DictID,
				"code":    old.Code,
				"name":    old.Name,
			},
		})

		refreshDictCache(ctx, tenantID, d.Code)
		return nil
	})
}

// refreshDictCache 失效/重建某字典的内存缓存
// ctx 走 WithoutCancel detach，避免调用方 cancel 阻断缓存重建
func refreshDictCache(parent context.Context, tenantID uint, code string) {
	if code == "" {
		return
	}
	ctx, cancel := context.WithCancel(context.WithoutCancel(parent))
	defer cancel()
	_ = dictpkg.RefreshDict(ctx, tenantID, code)
}

// validateAccess 校验 access 取值
func validateAccess(access string) error {
	switch access {
	case AccessInvisible, AccessReadonly, AccessEditable:
		return nil
	default:
		return ErrInvalidAccess
	}
}

// validateVisibility 校验 visibility 取值
func validateVisibility(v string) error {
	if v == "" {
		return nil
	}
	switch v {
	case VisibilityAll, VisibilityWhitelist, VisibilityBlacklist:
		return nil
	default:
		return ErrInvalidVisibility
	}
}

// ============ 平台字典 CRUD（仅 super_admin） ============

// ListPlatformDicts 平台字典列表（跨租户）
func (s *Service) ListPlatformDicts(ctx context.Context, req listRequest) ([]Dict, int64, error) {
	return s.repo.ListPlatformDicts(ctx, trimKeyword(req.Keyword), req.Page, req.Size)
}

// GetPlatformDict 平台字典详情
func (s *Service) GetPlatformDict(ctx context.Context, id uint) (*Dict, error) {
	return s.repo.GetPlatformDictByID(ctx, id)
}

// CreatePlatformDict 创建平台字典
func (s *Service) CreatePlatformDict(ctx context.Context, req platformDictCreateRequest) (*Dict, error) {
	if err := validateVisibility(req.Visibility); err != nil {
		return nil, err
	}
	// 把 visibility 字段写到 extend（也可以走单独列；这里走列）
	d, err := s.repo.CreatePlatformDict(ctx, CreateDictRepoReq{
		Code:   req.Code,
		Name:   req.Name,
		Sort:   req.Sort,
		Status: 1,
		Extend: req.Extend,
	})
	if err != nil {
		return nil, err
	}
	// 重新读取以应用 visibility 列（CreatePlatformDict 默认 all）
	if req.Visibility != "" && req.Visibility != d.Visibility {
		updReq := UpdateDictRepoReq{Name: d.Name, Sort: d.Sort, Status: d.Status, Extend: d.Extend}
		d2, uerr := s.repo.UpdatePlatformDict(ctx, d.ID, updReq)
		if uerr != nil {
			return nil, uerr
		}
		d = d2
		// 单独更新 visibility 列
		if _, err := s.pool.Exec(ctx, `UPDATE dicts SET visibility = $1 WHERE id = $2`, req.Visibility, d.ID); err != nil {
			return nil, fmt.Errorf("update visibility: %w", err)
		}
		d.Visibility = req.Visibility
	}

	// 审计：平台字典创建（高敏操作，跨租户影响，必须留痕）
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict:platform_create",
		TableName: "dicts",
		RecordID:  d.ID,
		NewData: map[string]any{
			"id":         d.ID,
			"code":       d.Code,
			"name":       d.Name,
			"sort":       d.Sort,
			"status":     d.Status,
			"scope":      d.Scope,
			"visibility": d.Visibility,
		},
	})
	return d, nil
}

// UpdatePlatformDict 更新平台字典
func (s *Service) UpdatePlatformDict(ctx context.Context, id uint, req platformDictUpdateRequest) (*Dict, error) {
	if err := validateVisibility(req.Visibility); err != nil {
		return nil, err
	}
	// 改前快照（用于审计 diff）
	before, err := s.repo.GetPlatformDictByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d, err := s.repo.UpdatePlatformDict(ctx, id, UpdateDictRepoReq{
		Name: req.Name, Sort: req.Sort, Status: req.Status, Extend: req.Extend,
	})
	if err != nil {
		return nil, err
	}
	if req.Visibility != "" && req.Visibility != d.Visibility {
		if _, err := s.pool.Exec(ctx, `UPDATE dicts SET visibility = $1 WHERE id = $2`, req.Visibility, id); err != nil {
			return nil, fmt.Errorf("update visibility: %w", err)
		}
		d.Visibility = req.Visibility
	}

	// 审计：平台字典更新（高敏操作，跨租户影响）
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict:platform_update",
		TableName: "dicts",
		RecordID:  id,
		OldData: map[string]any{
			"name":       before.Name,
			"sort":       before.Sort,
			"status":     before.Status,
			"visibility": before.Visibility,
		},
		NewData: map[string]any{
			"name":       d.Name,
			"sort":       d.Sort,
			"status":     d.Status,
			"visibility": d.Visibility,
		},
	})
	return d, nil
}

// DeletePlatformDict 删除平台字典（仍有租户覆盖时拒绝）
func (s *Service) DeletePlatformDict(ctx context.Context, id uint) error {
	// 改前快照（审计用）
	before, err := s.repo.GetPlatformDictByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeletePlatformDict(ctx, id); err != nil {
		return err
	}

	// 审计：平台字典删除（不可逆高敏操作）
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict:platform_delete",
		TableName: "dicts",
		RecordID:  id,
		OldData: map[string]any{
			"id":         before.ID,
			"code":       before.Code,
			"name":       before.Name,
			"scope":      before.Scope,
			"visibility": before.Visibility,
			"status":     before.Status,
		},
	})
	return nil
}

// ============ 平台字典项 CRUD（仅 super_admin） ============

// ListPlatformItems 平台字典项
func (s *Service) ListPlatformItems(ctx context.Context, dictID uint) ([]DictItem, error) {
	// 先校验 dict 是 platform
	d, err := s.repo.GetPlatformDictByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	_ = d
	return s.repo.ListPlatformItems(ctx, dictID)
}

// CreatePlatformItem 新增平台字典项
func (s *Service) CreatePlatformItem(ctx context.Context, dictID uint, req platformItemCreateRequest) (*DictItem, error) {
	d, err := s.repo.GetPlatformDictByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.CreatePlatformItem(ctx, dictID, CreateDictItemRepoReq{
		Code: req.Code, Name: req.Name, Sort: req.Sort, Status: 1, Extend: req.Extend,
	})
	if err != nil {
		return nil, err
	}

	// 审计：平台字典项新增
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict_item:platform_create",
		TableName: "dict_items",
		RecordID:  item.ID,
		NewData: map[string]any{
			"id":      item.ID,
			"dict_id": item.DictID,
			"dict":    d.Code,
			"code":    item.Code,
			"name":    item.Name,
			"sort":    item.Sort,
			"status":  item.Status,
		},
	})
	return item, nil
}

// UpdatePlatformItem 更新平台字典项
func (s *Service) UpdatePlatformItem(ctx context.Context, itemID uint, req updateItemRequest) error {
	status := req.Status
	if status == 0 {
		status = 1
	}
	// 改前快照
	before, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if before.TenantID != 0 {
		return ErrDictItemNotFound // 只允许改平台项
	}
	if err := s.repo.UpdatePlatformItem(ctx, itemID, UpdateDictItemRepoReq{
		Name: req.Name, Sort: req.Sort, Status: status, Extend: req.Extend,
	}); err != nil {
		return err
	}

	// 审计：平台字典项更新
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict_item:platform_update",
		TableName: "dict_items",
		RecordID:  itemID,
		OldData: map[string]any{
			"name":   before.Name,
			"sort":   before.Sort,
			"status": before.Status,
			"code":   before.Code,
		},
		NewData: map[string]any{
			"name":   req.Name,
			"sort":   req.Sort,
			"status": status,
		},
	})
	return nil
}

// DeletePlatformItem 删除平台字典项
func (s *Service) DeletePlatformItem(ctx context.Context, itemID uint) error {
	// 改前快照
	before, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if before.TenantID != 0 {
		return ErrDictItemNotFound
	}
	if err := s.repo.DeletePlatformItem(ctx, itemID); err != nil {
		return err
	}

	// 审计：平台字典项删除（不可逆）
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict_item:platform_delete",
		TableName: "dict_items",
		RecordID:  itemID,
		OldData: map[string]any{
			"id":      before.ID,
			"dict_id": before.DictID,
			"code":    before.Code,
			"name":    before.Name,
		},
	})
	return nil
}

// ============ 可见性配置（仅 super_admin） ============

// ListVisibility 平台字典对各租户的可见性列表
func (s *Service) ListVisibility(ctx context.Context, dictID uint) ([]DictVisibility, error) {
	if _, err := s.repo.GetPlatformDictByID(ctx, dictID); err != nil {
		return nil, err
	}
	return s.repo.ListVisibilityByDict(ctx, dictID)
}

// UpsertVisibility upsert 单条可见性配置
func (s *Service) UpsertVisibility(ctx context.Context, dictID uint, req visibilityUpsertRequest) (*DictVisibility, error) {
	if err := validateAccess(req.Access); err != nil {
		return nil, err
	}
	d, err := s.repo.GetPlatformDictByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	// 改前快照（用于 audit diff；可能不存在）
	before, _ := s.repo.GetAccessForTenant(ctx, dictID, req.TenantID)

	v, err := s.repo.UpsertVisibility(ctx, dictID, req.TenantID, req.Access)
	if err != nil {
		return nil, err
	}

	// 审计：可见性配置变更（影响特定租户的访问权限）
	action := "dict_visibility:upsert"
	if before != "" && before != req.Access {
		action = "dict_visibility:update"
	} else if before == "" {
		action = "dict_visibility:create"
	}
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    action,
		TableName: "dict_visibility",
		RecordID:  v.ID,
		OldData: map[string]any{
			"dict_id":   dictID,
			"dict":      d.Code,
			"tenant_id": req.TenantID,
			"access":    before,
		},
		NewData: map[string]any{
			"dict_id":   dictID,
			"dict":      d.Code,
			"tenant_id": req.TenantID,
			"access":    req.Access,
		},
	})
	return v, nil
}

// DeleteVisibility 删除单条可见性配置
func (s *Service) DeleteVisibility(ctx context.Context, dictID, tenantID uint) error {
	// 改前快照
	d, err := s.repo.GetPlatformDictByID(ctx, dictID)
	if err != nil {
		return err
	}
	before, _ := s.repo.GetAccessForTenant(ctx, dictID, tenantID)
	if err := s.repo.DeleteVisibility(ctx, dictID, tenantID); err != nil {
		return err
	}

	// 审计：可见性配置删除（恢复默认策略）
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict_visibility:delete",
		TableName: "dict_visibility",
		RecordID:  dictID,
		OldData: map[string]any{
			"dict_id":   dictID,
			"dict":      d.Code,
			"tenant_id": tenantID,
			"access":    before,
		},
	})
	return nil
}

// ============ 租户覆盖（override） ============

// UpsertOverride 租户对平台字典项 upsert 覆盖
func (s *Service) UpsertOverride(ctx context.Context, tenantID, dictID, platformItemID uint, req overrideUpsertRequest) (*DictItem, error) {
	// 校验 dict 是 platform
	d, err := s.repo.GetPlatformDictByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	// 校验 access=editable
	access, err := s.repo.GetAccessForTenant(ctx, d.ID, tenantID)
	if err != nil {
		return nil, err
	}
	if access == "" {
		// 走默认策略
		switch d.Visibility {
		case VisibilityWhitelist:
			return nil, ErrDictInvisible
		default:
			access = AccessEditable
		}
	}
	if access != AccessEditable {
		return nil, ErrDictReadonly
	}

	// 改前快照（用于审计 diff）
	before, _ := s.repo.GetOverrideByPlatformItem(ctx, platformItemID, tenantID)

	status := req.Status
	if status == 0 {
		status = 1
	}
	item, err := s.repo.UpsertOverride(ctx, tenantID, dictID, platformItemID, UpdateDictItemRepoReq{
		Name: req.Name, Sort: req.Sort, Status: status, Extend: req.Extend,
	})
	if err != nil {
		return nil, err
	}
	// 失效缓存
	refreshDictCache(ctx, tenantID, d.Code)

	// 审计：租户覆盖字典项（create / update 区分）
	action := "dict_item:override_upsert"
	if before == nil {
		action = "dict_item:override_create"
	} else {
		action = "dict_item:override_update"
	}
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    action,
		TableName: "dict_items",
		RecordID:  item.ID,
		TenantID:  tenantID, // 显式覆盖（覆盖属于租户域操作）
		OldData: func() map[string]any {
			if before == nil {
				return map[string]any{
					"platform_item_id": platformItemID,
					"dict":             d.Code,
				}
			}
			return map[string]any{
				"name":             before.Name,
				"sort":             before.Sort,
				"status":           before.Status,
				"platform_item_id": platformItemID,
				"dict":             d.Code,
			}
		}(),
		NewData: map[string]any{
			"name":             item.Name,
			"sort":             item.Sort,
			"status":           item.Status,
			"platform_item_id": platformItemID,
			"dict":             d.Code,
			"is_override":      item.IsOverride,
		},
	})
	return item, nil
}

// DeleteOverride 取消租户覆盖
func (s *Service) DeleteOverride(ctx context.Context, tenantID, dictID, platformItemID uint) error {
	d, err := s.repo.GetPlatformDictByID(ctx, dictID)
	if err != nil {
		return err
	}
	// 改前快照
	before, _ := s.repo.GetOverrideByPlatformItem(ctx, platformItemID, tenantID)
	if err := s.repo.DeleteOverride(ctx, tenantID, platformItemID); err != nil {
		return err
	}
	refreshDictCache(ctx, tenantID, d.Code)

	// 审计：取消租户覆盖（恢复平台默认值）
	audit.Log(ctx, s.pool, audit.Entry{
		Action:    "dict_item:override_delete",
		TableName: "dict_items",
		RecordID:  dictID,
		TenantID:  tenantID,
		OldData: func() map[string]any {
			out := map[string]any{
				"platform_item_id": platformItemID,
				"dict":             d.Code,
			}
			if before != nil {
				out["name"] = before.Name
				out["sort"] = before.Sort
				out["status"] = before.Status
				out["override_id"] = before.ID
			}
			return out
		}(),
	})
	return nil
}

// ============ 业务消费：Resolve 合并字典 ============

// ResolveForTenant 按 code 取租户视角的合并字典
func (s *Service) ResolveForTenant(ctx context.Context, tenantID uint, dictCode string) (*ResolvedDict, error) {
	return s.repo.ResolveDictForTenant(ctx, tenantID, dictCode)
}

// ResolveByIDForTenant 按 dict_id 取租户视角的合并字典
func (s *Service) ResolveByIDForTenant(ctx context.Context, tenantID, dictID uint) (*ResolvedDict, error) {
	return s.repo.ResolveDictByIDForTenant(ctx, tenantID, dictID)
}

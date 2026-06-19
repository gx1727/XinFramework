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

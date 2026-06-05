// Package dict ?????
package dict

import (
	"context"
	"fmt"
	"strings"

	"gx1727.com/xin/framework/pkg/audit"
	"gx1727.com/xin/framework/pkg/db"
	dictpkg "gx1727.com/xin/framework/pkg/dict"
)

type Service struct {
	repo DictRepository
}

func NewService(repo DictRepository) *Service {
	return &Service{repo: repo}
}

// ========== ???? ==========

// List ????
func (s *Service) List(ctx context.Context, tenantID uint, req listRequest) ([]Dict, int64, error) {
	return s.repo.List(ctx, tenantID, strings.TrimSpace(req.Keyword), req.Page, req.Size)
}

// Get ??????
func (s *Service) Get(ctx context.Context, tenantID uint, id uint) (*Dict, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.TenantID != tenantID {
		return nil, ErrDictNotFound
	}
	return d, nil
}

// Create ????
func (s *Service) Create(ctx context.Context, tenantID uint, req createRequest) (*Dict, error) {
	return s.repo.Create(ctx, tenantID, CreateDictRepoReq{
		Code:   req.Code,
		Name:   req.Name,
		Sort:   req.Sort,
		Status: 1,
		Extend: req.Extend,
	})
}

// Update ????
func (s *Service) Update(ctx context.Context, tenantID, id uint, req updateRequest) (*Dict, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.TenantID != tenantID {
		return nil, ErrDictNotFound
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	return s.repo.Update(ctx, id, UpdateDictRepoReq{
		Name:   req.Name,
		Sort:   req.Sort,
		Status: status,
		Extend: req.Extend,
	})
}

// Delete ???????????????????????
func (s *Service) Delete(ctx context.Context, tenantID, id uint) error {
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
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
			return fmt.Errorf("%w (????=%d)", ErrDictHasItems, n)
		}

		if err := s.repo.Delete(ctx, id); err != nil {
			return err
		}

		// ???????
		audit.Log(ctx, audit.Entry{
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

// ========== ??? ==========

// ListItems ????????????
func (s *Service) ListItems(ctx context.Context, tenantID, dictID uint) ([]DictItem, error) {
	d, err := s.repo.GetByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	if d.TenantID != tenantID {
		return nil, ErrDictNotFound
	}
	return s.repo.ListItems(ctx, dictID)
}

// CreateItem ????????????? + ????
func (s *Service) CreateItem(ctx context.Context, tenantID, dictID uint, req createItemRequest) (*DictItem, error) {
	d, err := s.repo.GetByID(ctx, dictID)
	if err != nil {
		return nil, err
	}
	if d.TenantID != tenantID {
		return nil, ErrDictNotFound
	}

	item, err := s.repo.CreateItem(ctx, tenantID, dictID, CreateDictItemRepoReq{
		Code:   req.Code,
		Name:   req.Name,
		Sort:   req.Sort,
		Status: 1,
		Extend: req.Extend,
	})
	if err != nil {
		return nil, err
	}

	refreshDictCache(ctx, tenantID, d.Code)
	return item, nil
}

// UpdateItem ????????? + ????
func (s *Service) UpdateItem(ctx context.Context, tenantID, itemID uint, req updateItemRequest) error {
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
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

		audit.Log(ctx, audit.Entry{
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

// DeleteItem ????????? + ????
func (s *Service) DeleteItem(ctx context.Context, tenantID, itemID uint) error {
	return db.RunInTenantTx(ctx, db.Get(), tenantID, func(ctx context.Context) error {
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

		audit.Log(ctx, audit.Entry{
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

// refreshDictCache ??/???????????
// ??? ctx ?? cancel??? detach ??????????
func refreshDictCache(parent context.Context, tenantID uint, code string) {
	if code == "" {
		return
	}
	ctx, cancel := context.WithCancel(context.WithoutCancel(parent))
	defer cancel()
	_ = dictpkg.RefreshDict(ctx, tenantID, code)
}

// Package syspermission 实现"sys 域权限码"管理 API（sys_permissions 表）。
package syspermission

import (
	"context"
	"time"
)

type Permission struct {
	ID          uint      `json:"id"`
	MenuID      *uint     `json:"menu_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRepoReq struct {
	MenuID      *uint
	Code        string
	Name        string
	Action      string
	Description string
	Sort        int
	Status      int8
	CreatedBy   uint
}

type UpdateRepoReq struct {
	MenuID      *uint
	Code        *string
	Name        *string
	Action      *string
	Description *string
	Sort        *int
	Status      *int8
	UpdatedBy   uint
}

type Repository interface {
	GetByID(ctx context.Context, id uint) (*Permission, error)
	GetByCode(ctx context.Context, code string) ([]Permission, error)
	List(ctx context.Context, menuID *uint, keyword string, page, size int) ([]Permission, int64, error)
	Create(ctx context.Context, req CreateRepoReq) (*Permission, error)
	Update(ctx context.Context, id uint, req UpdateRepoReq) (*Permission, error)
	Delete(ctx context.Context, id uint, updatedBy uint) error
}

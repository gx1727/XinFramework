// Package sysorg 实现"sys 域组织"管理 API（sys_orgs 表）。
//
// Phase 0023.0 简版：
//   - 仅暴露 model/types/errors/repository（只读 GetByID + List）
//   - 不挂 HTTP 路由（不写 handler.go / service.go / routes.go）
//   - 不写 module.go（在主程序里不注册为 plugin module）
//
// 为什么先不做完整 CRUD？
//   - sys_orgs 的业务模型需要先和 sys_users、sys_role_data_scopes 联合验证
//   - 业务方当前没有强需求（super_admin 是单点管理，暂无层级组织）
//   - YAGNI：等到需要"按组织限定 sys 管理员权限"时再补
//
// 未来 Phase 0023.1 补全时的注意点：
//   - sys_orgs ancestors 用 parent_id 闭包计算（参考 tenant_orgs）
//   - 删除时需要禁止：sys_users.org_id 仍引用
package sysorg

import (
	"context"
	"time"
)

type Org struct {
	ID          uint      `json:"id"`
	ParentID    *uint     `json:"parent_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AdminCode   string    `json:"admin_code"`
	Ancestors   string    `json:"ancestors"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository interface {
	GetByID(ctx context.Context, id uint) (*Org, error)
	List(ctx context.Context, keyword string, page, size int) ([]Org, int64, error)
	// Phase 0023.1 补：
	//   Create / Update / Delete
	//   GetTree
}

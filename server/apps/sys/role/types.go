package sysrole

import "time"

type CreateSysRoleReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	OrgID       *uint  `json:"org_id"`
	DataScope   int8   `json:"data_scope"`
	IsDefault   bool   `json:"is_default"`
	Sort        int    `json:"sort"`
	Status      *int8  `json:"status"`
	MenuIDs     []uint `json:"menu_ids"`
	PermIDs     []uint `json:"permission_ids"`
}

type UpdateSysRoleReq struct {
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	OrgID       *uint   `json:"org_id"`
	DataScope   *int8   `json:"data_scope"`
	IsDefault   *bool   `json:"is_default"`
	Sort        *int    `json:"sort"`
	Status      *int8   `json:"status"`
}

type ListQuery struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type AssignMenusReq struct {
	MenuIDs []uint `json:"menu_ids" binding:"required"`
}

type AssignPermissionsReq struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

type SysRoleResp struct {
	ID          uint             `json:"id"`
	OrgID       *uint            `json:"org_id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	DataScope   int8             `json:"data_scope"`
	IsDefault   bool             `json:"is_default"`
	Sort        int              `json:"sort"`
	Status      int8             `json:"status"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	Menus       []MenuLite       `json:"menus,omitempty"`
	Permissions []PermissionLite `json:"permissions,omitempty"`
}

func toResp(r *Role) *SysRoleResp {
	if r == nil {
		return nil
	}
	return &SysRoleResp{
		ID:          r.ID,
		OrgID:       r.OrgID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		DataScope:   r.DataScope,
		IsDefault:   r.IsDefault,
		Sort:        r.Sort,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.Format(time.RFC3339),
	}
}

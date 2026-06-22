package syspermission

import "time"

type CreateSysPermissionReq struct {
	MenuID      *uint  `json:"menu_id"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      *int8  `json:"status"`
}

type UpdateSysPermissionReq struct {
	MenuID      *uint   `json:"menu_id"`
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	Action      *string `json:"action"`
	Description *string `json:"description"`
	Sort        *int    `json:"sort"`
	Status      *int8   `json:"status"`
}

type ListQuery struct {
	MenuID  *uint  `form:"menu_id"`
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type SysPermissionResp struct {
	ID          uint     `json:"id"`
	MenuID      *uint    `json:"menu_id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Action      string   `json:"action"`
	Description string   `json:"description"`
	Sort        int      `json:"sort"`
	Status      int8     `json:"status"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	MenuCode    *string  `json:"menu_code,omitempty"`
}

func toResp(p *Permission) *SysPermissionResp {
	if p == nil {
		return nil
	}
	return &SysPermissionResp{
		ID:          p.ID,
		MenuID:      p.MenuID,
		Code:        p.Code,
		Name:        p.Name,
		Action:      p.Action,
		Description: p.Description,
		Sort:        p.Sort,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

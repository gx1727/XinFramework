package sysorg

import "time"

type ListQuery struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type SysOrgResp struct {
	ID          uint     `json:"id"`
	ParentID    *uint    `json:"parent_id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	AdminCode   string   `json:"admin_code"`
	Ancestors   string   `json:"ancestors"`
	Sort        int      `json:"sort"`
	Status      int8     `json:"status"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func toResp(o *Org) *SysOrgResp {
	if o == nil {
		return nil
	}
	return &SysOrgResp{
		ID:          o.ID,
		ParentID:    o.ParentID,
		Code:        o.Code,
		Name:        o.Name,
		Type:        o.Type,
		Description: o.Description,
		AdminCode:   o.AdminCode,
		Ancestors:   o.Ancestors,
		Sort:        o.Sort,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   o.UpdatedAt.Format(time.RFC3339),
	}
}

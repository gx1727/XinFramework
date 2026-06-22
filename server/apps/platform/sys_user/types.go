package sysuser

import "time"

// CreateSysUserReq API 入参。
type CreateSysUserReq struct {
	AccountID uint   `json:"account_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
	OrgID     *uint  `json:"org_id"`
	RealName  string `json:"real_name" binding:"required"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Status    *int8  `json:"status"`
	RoleIDs   []uint `json:"role_ids"` // 可选：创建时直接绑角色
}

type UpdateSysUserReq struct {
	Code     *string `json:"code"`
	OrgID    *uint   `json:"org_id"`
	RealName *string `json:"real_name"`
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Status   *int8   `json:"status"`
}

type ListQuery struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type AssignRolesReq struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

// SysUserResp 返回结构。
type SysUserResp struct {
	ID        uint       `json:"id"`
	AccountID uint       `json:"account_id"`
	OrgID     *uint      `json:"org_id"`
	Code      string     `json:"code"`
	RealName  string     `json:"real_name"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Status    int8       `json:"status"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Roles     []RoleLite `json:"roles,omitempty"`
}

func toResp(u *User) *SysUserResp {
	if u == nil {
		return nil
	}
	return &SysUserResp{
		ID:        u.ID,
		AccountID: u.AccountID,
		OrgID:     u.OrgID,
		Code:      u.Code,
		RealName:  u.RealName,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		Roles:     u.Roles,
	}
}

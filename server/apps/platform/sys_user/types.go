package sysuser

import "time"

// CreateSysUserReq API 入参。
//
// 支持两种模式：
//
//  1. 绑定已有 account（兼容旧接口）：传 account_id > 0。
//     此时 phone/password/email 字段被忽略。
//
//  2. 一并新建可登录账号（AccountID 留 0）：必须给 phone + password。
//     username 留空时默认同 phone；email 选填；code 留空时后端按
//     "u<account_id>" 自动生成（≤32 字符）。
type CreateSysUserReq struct {
	// 模式 1：绑定已有账号
	AccountID uint `json:"account_id"`

	// 模式 2：新建可登录账号（AccountID == 0 时启用）
	Username string `json:"username"`
	Phone    string `json:"phone" binding:"required_without=AccountID"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required_without=AccountID,min=6,max=32"`

	// 平台身份字段
	Code     string `json:"code"`
	OrgID    *uint  `json:"org_id"`
	RealName string `json:"real_name" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   *int8  `json:"status"`
	RoleIDs  []uint `json:"role_ids"` // 可选：创建时直接绑角色
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

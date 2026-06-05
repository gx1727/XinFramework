package user

type listRequest struct {
	Keyword string `form:"keyword"`
	OrgID   *uint  `form:"org_id"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

type listResponse struct {
	List  []UserInfo `json:"list"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	TenantID  uint   `json:"tenant_id"`
	AccountID uint   `json:"account_id"`
	OrgID     *uint  `json:"org_id,omitempty"`
	OrgName   string `json:"org_name,omitempty"`
	Code      string `json:"code"`
	Nickname  string `json:"nickname"`
	RealName  string `json:"real_name"`
	Avatar    string `json:"avatar"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    int8   `json:"status"`
}

func (u *UserInfo) GetDisplayName() string {
	if u.Nickname != "" {
		return u.Nickname
	}
	if u.RealName != "" {
		return u.RealName
	}
	return u.Code
}

type getRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type updateStatusRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int8 `json:"status" binding:"required,oneof=1 2"`
}

type updateProfileRequest struct {
	Nickname string `json:"nickName" binding:"required"`
	Avatar   string `json:"avatarUrl"`
}

type createRequest struct {
	Username string `json:"username" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email"`
	RealName string `json:"real_name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	OrgID    *uint  `json:"org_id"`
	Status   int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

// updateUserRequest 全量替换；仅修改 users 表字段。
// phone/email 在 accounts 表上，由专门的"换绑手机/邮箱"流程处理。
type updateUserRequest struct {
	Nickname string `json:"nickname"`
	RealName string `json:"real_name"`
	Avatar   string `json:"avatar"`
	OrgID    *uint  `json:"org_id"`
	Status   int8   `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// patchUserRequest 局部更新；nil 字段表示保持原值
type patchUserRequest struct {
	Nickname *string `json:"nickname"`
	RealName *string `json:"real_name"`
	Avatar   *string `json:"avatar"`
	OrgID    *uint   `json:"org_id"`
	Status   *int8   `json:"status" binding:"omitempty,oneof=0 1 2"`
}

type createResponse struct {
	ID       uint   `json:"id"`
	TenantID uint   `json:"tenant_id"`
	Code     string `json:"code"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	OrgID    *uint  `json:"org_id,omitempty"`
	OrgName  string `json:"org_name,omitempty"`
	Status   int8   `json:"status"`
}

// updateOrgRequest 用于 PUT /users/:id/org 调整主组织。
// 传 0 表示把用户移出组织（org_id 置 NULL），不传/传 null 同效。
type updateOrgRequest struct {
	OrgID *uint `json:"org_id"`
}

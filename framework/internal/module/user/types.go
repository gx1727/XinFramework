package user

type listRequest struct {
	Keyword string `form:"keyword"`
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
	Code      string `json:"code"`
	Nickname  string `json:"nickname"`
	RealName  string `json:"real_name"`
	Avatar    string `json:"avatar"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Role      string `json:"role"`
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
	Status   int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

type createResponse struct {
	ID       uint   `json:"id"`
	TenantID uint   `json:"tenant_id"`
	Code     string `json:"code"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
}

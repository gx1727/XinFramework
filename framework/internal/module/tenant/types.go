package tenant

type CreateTenantReq struct {
	Code    string `json:"code" binding:"required,min=1,max=50"`
	Name    string `json:"name" binding:"required,min=1,max=100"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Status  *int16 `json:"status"`
}

type UpdateTenantReq struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Contact  string `json:"contact"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   *int16 `json:"status"`
	Province string `json:"province"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Address  string `json:"address"`
}

type ListTenantReq struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Size    int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword"`
	Status  *int16 `form:"status"`
}

type TenantResp struct {
	ID        uint   `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Status    int16  `json:"status"`
	Contact   string `json:"contact"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Area      string `json:"area"`
	Address   string `json:"address"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

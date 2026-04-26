package resource

type ListReq struct {
	MenuID uint   `form:"menu_id"`
	Action string `form:"action"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=20"`
}

type CreateReq struct {
	MenuID      uint   `json:"menu_id"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type UpdateReq struct {
	Name        string `json:"name" binding:"required"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type ResourceResp struct {
	ID          uint   `json:"id"`
	TenantID    uint   `json:"tenant_id"`
	MenuID      uint   `json:"menu_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type listResponse struct {
	List  []ResourceResp `json:"list"`
	Total int64          `json:"total"`
}

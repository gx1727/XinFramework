package role

type ListReq struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

type CreateReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	DataScope   int8   `json:"data_scope"`
	IsDefault   bool   `json:"is_default"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type UpdateReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	DataScope   int8   `json:"data_scope"`
	IsDefault   bool   `json:"is_default"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type UpdateDataScopesReq struct {
	OrgIDs []uint `json:"org_ids"`
}

type RoleResp struct {
	ID          uint   `json:"id"`
	TenantID    uint   `json:"tenant_id"`
	OrgID       uint   `json:"org_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DataScope   int8   `json:"data_scope"`
	Extend      string `json:"extend"`
	IsDefault   bool   `json:"is_default"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type DataScopeResp struct {
	OrgIDs []uint `json:"org_ids"`
}

type listResponse struct {
	List  []RoleResp `json:"list"`
	Total int64      `json:"total"`
}

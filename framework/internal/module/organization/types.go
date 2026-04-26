package organization

type ListReq struct {
	Keyword  string `form:"keyword"`
	ParentID uint   `form:"parent_id"`
	Page     int    `form:"page,default=1"`
	Size     int    `form:"size,default=20"`
}

type CreateReq struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
	AdminCode   string `json:"admin_code"`
	ParentID    uint   `json:"parent_id"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type UpdateReq struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
	AdminCode   string `json:"admin_code"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

type OrgResp struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AdminCode   string    `json:"admin_code"`
	ParentID    uint      `json:"parent_id"`
	Ancestors   string    `json:"ancestors"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Children    []OrgResp `json:"children,omitempty"`
}

type listResponse struct {
	List  []OrgResp `json:"list"`
	Total int64     `json:"total"`
}

type treeResponse struct {
	Tree []OrgResp `json:"tree"`
}

package menu

type CreateMenuReq struct {
	Code      string `json:"code" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Subtitle  string `json:"subtitle"`
	URL       string `json:"url"`
	Path      string `json:"path"`
	Icon      string `json:"icon"`
	Sort      int    `json:"sort"`
	ParentID  uint   `json:"parent_id"`
	Ancestors string `json:"ancestors"`
	Visible   *bool  `json:"visible"`
	Enabled   *bool  `json:"enabled"`
}

type UpdateMenuReq struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
	Visible  *bool  `json:"visible"`
	Enabled  *bool  `json:"enabled"`
}

type ListMenuReq struct {
	Page int  `form:"page"`
	Size int  `form:"size"`
	Root bool `form:"root"` // 只查顶级菜单
}

type MenuResp struct {
	ID        uint        `json:"id"`
	TenantID  uint        `json:"tenant_id"`
	Code      string      `json:"code"`
	Name      string      `json:"name"`
	Subtitle  string      `json:"subtitle"`
	URL       string      `json:"url"`
	Path      string      `json:"path"`
	Icon      string      `json:"icon"`
	Sort      int         `json:"sort"`
	ParentID  uint        `json:"parent_id"`
	Ancestors string      `json:"ancestors"`
	Visible   bool        `json:"visible"`
	Enabled   bool        `json:"enabled"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	Children  []*MenuResp `json:"children,omitempty"`
}

type TreeNode struct {
	Menu     *MenuResp
	Children []*TreeNode
}

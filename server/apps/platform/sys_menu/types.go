package sysmenu

import "time"

type CreateSysMenuReq struct {
	Code      string  `json:"code" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Subtitle  *string `json:"subtitle"`
	URL       *string `json:"url"`
	Path      *string `json:"path"`
	Icon      *string `json:"icon"`
	Sort      int     `json:"sort"`
	ParentID  *uint   `json:"parent_id"`
	Ancestors *string `json:"ancestors"`
	Visible   *bool   `json:"visible"`
	Enabled   *bool   `json:"enabled"`
}

type UpdateSysMenuReq struct {
	Code      *string `json:"code"`
	Name      *string `json:"name"`
	Subtitle  *string `json:"subtitle"`
	URL       *string `json:"url"`
	Path      *string `json:"path"`
	Icon      *string `json:"icon"`
	Sort      *int    `json:"sort"`
	ParentID  *uint   `json:"parent_id"`
	Ancestors *string `json:"ancestors"`
	Visible   *bool   `json:"visible"`
	Enabled   *bool   `json:"enabled"`
}

type SysMenuResp struct {
	ID        uint       `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Subtitle  *string    `json:"subtitle"`
	URL       *string    `json:"url"`
	Path      *string    `json:"path"`
	Icon      *string    `json:"icon"`
	Sort      int        `json:"sort"`
	ParentID  *uint      `json:"parent_id"`
	Ancestors *string    `json:"ancestors"`
	Visible   bool       `json:"visible"`
	Enabled   bool       `json:"enabled"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Children  []*SysMenuResp `json:"children,omitempty"`
}

func toResp(m *Menu) *SysMenuResp {
	if m == nil {
		return nil
	}
	return &SysMenuResp{
		ID:        m.ID,
		Code:      m.Code,
		Name:      m.Name,
		Subtitle:  m.Subtitle,
		URL:       m.URL,
		Path:      m.Path,
		Icon:      m.Icon,
		Sort:      m.Sort,
		ParentID:  m.ParentID,
		Ancestors: m.Ancestors,
		Visible:   m.Visible,
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func buildTree(menus []Menu) []*SysMenuResp {
	nodeMap := make(map[uint]*SysMenuResp, len(menus))
	for i := range menus {
		nodeMap[menus[i].ID] = toResp(&menus[i])
	}
	var roots []*SysMenuResp
	for i := range menus {
		m := &menus[i]
		node := nodeMap[m.ID]
		if m.ParentID == nil || *m.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodeMap[*m.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	return roots
}

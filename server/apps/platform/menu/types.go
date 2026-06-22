package menu

import "time"

// CreateMenuReq 平台菜单创建请求。
// 与 rbac/menu.CreateMenuReq 字段一致，但绑定的 code 走"平台级"
//
// 校验（不能与已有平台菜单 code 重复，也不能与租户菜单 code 重复：
// 因为 uk_menu_code (tenant_id, code) 唯一，跨 tenant_id 不冲突，
// 所以这里只校验平台级内部唯一）。
type CreateMenuReq struct {
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

type UpdateMenuReq struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
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

// MenuResp 平台菜单返回结构。Children 在 tree 接口填充。
type MenuResp struct {
	ID        uint        `json:"id"`
	TenantID  uint        `json:"tenant_id"` // 永远 = 0
	Code      string      `json:"code"`
	Name      string      `json:"name"`
	Subtitle  *string     `json:"subtitle"`
	URL       *string     `json:"url"`
	Path      *string     `json:"path"`
	Icon      *string     `json:"icon"`
	Sort      int         `json:"sort"`
	ParentID  *uint       `json:"parent_id"`
	Ancestors *string     `json:"ancestors"`
	Visible   bool        `json:"visible"`
	Enabled   bool        `json:"enabled"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	Children  []*MenuResp `json:"children,omitempty"`
}

func toResp(m *Menu) *MenuResp {
	if m == nil {
		return nil
	}
	return &MenuResp{
		ID:        m.ID,
		TenantID:  m.TenantID,
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

// buildTree 从扁平列表构造树。
//
// 与 rbac/menu.buildTree 同样的算法（DFS cycle detection），
// 这里是独立实现，避免跨包导出内部函数。
func buildTree(menus []Menu) []*MenuResp {
	nodeMap := make(map[uint]*MenuResp, len(menus))
	for i := range menus {
		nodeMap[menus[i].ID] = toResp(&menus[i])
	}

	var roots []*MenuResp
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

	// cycle break：DFS 检测回边
	visited := make(map[uint]int) // 0=unvisited, 1=visiting, 2=done
	var breakCycle func(nodeID uint) bool
	breakCycle = func(nodeID uint) bool {
		if nodeID == 0 {
			return false
		}
		state, ok := visited[nodeID]
		if !ok {
			visited[nodeID] = 1
		} else if state == 1 {
			if node, exists := nodeMap[nodeID]; exists {
				removeFromParent(node, roots)
				if !inRoots(node, roots) {
					roots = append(roots, node)
				}
			}
			return true
		} else {
			return false
		}
		if node, ok := nodeMap[nodeID]; ok && node.ParentID != nil && *node.ParentID != 0 {
			breakCycle(*node.ParentID)
		}
		visited[nodeID] = 2
		return false
	}

	for i := range menus {
		breakCycle(menus[i].ID)
	}
	return roots
}

func removeFromParent(node *MenuResp, roots []*MenuResp) {
	for _, root := range roots {
		for j := len(root.Children) - 1; j >= 0; j-- {
			if root.Children[j].ID == node.ID {
				root.Children = append(root.Children[:j], root.Children[j+1:]...)
				return
			}
			removeFromParentDeep(root.Children[j], node, roots)
		}
	}
}

func removeFromParentDeep(parent *MenuResp, target *MenuResp, roots []*MenuResp) {
	for j := len(parent.Children) - 1; j >= 0; j-- {
		if parent.Children[j].ID == target.ID {
			parent.Children = append(parent.Children[:j], parent.Children[j+1:]...)
			return
		}
		removeFromParentDeep(parent.Children[j], target, roots)
	}
}

func inRoots(node *MenuResp, roots []*MenuResp) bool {
	for _, r := range roots {
		if r.ID == node.ID {
			return true
		}
	}
	return false
}

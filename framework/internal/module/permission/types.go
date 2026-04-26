package permission

type AssignReq struct {
	Permissions []PermissionAssign `json:"permissions" binding:"required"`
}

type PermissionAssign struct {
	ResourceType string `json:"resource_type" binding:"required"` // menu, resource, route
	ResourceID   uint   `json:"resource_id"`
	ResourceCode string `json:"resource_code"`
	Effect       int8   `json:"effect"` // 1=allow, 0=deny
}

type RolePermissionsResp struct {
	Menus     []MenuPerm     `json:"menus"`
	Resources []ResourcePerm `json:"resources"`
}

type MenuPerm struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Effect int8   `json:"effect"`
}

type ResourcePerm struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Action string `json:"action"`
	Effect int8   `json:"effect"`
}

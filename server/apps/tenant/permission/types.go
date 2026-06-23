package permission

// AssignResourceReq 分配角色资源权限请求（全量覆盖）
type AssignResourceReq struct {
	ResourceIDs []uint `json:"resource_ids" binding:"required"`
}

// ResourcePerm 角色已分配的资源权限
type ResourcePerm struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Action string `json:"action"`
}

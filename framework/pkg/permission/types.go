package permission

// DataScopeType defines the type of data access scope
type DataScopeType int

const (
	DataScopeAll          DataScopeType = 1 // 全部数据
	DataScopeCustom       DataScopeType = 2 // 自定义数据
	DataScopeDept         DataScopeType = 3 // 本部门数据
	DataScopeDeptAndBelow DataScopeType = 4 // 本部门及以下数据
	DataScopeSelf         DataScopeType = 5 // 本人数据
)

// DataScope represents data access scope for a user
type DataScope struct {
	Type   DataScopeType `json:"type"`
	OrgIDs []int64       `json:"org_ids,omitempty"` // for type=2 (custom)
}

// Permission represents a role permission record
type Permission struct {
	ID           uint
	TenantID     uint
	RoleID       uint
	ResourceType string
	ResourceID   uint
	ResourceCode string
	Effect       int8
}

// HasPermission checks if a permission map contains the given permission
// Supports wildcard matching: "user:*" grants all actions on user resource,
// "*:*" grants all permissions (super admin)
func HasPermission(perms map[string]bool, resource, action string) bool {
	if perms == nil {
		return false
	}

	key := resource + ":" + action

	// Check exact match
	if perms[key] {
		return true
	}

	// Check wildcard for the resource (e.g., "user:*" grants "user:delete")
	if perms[resource+":*"] {
		return true
	}

	// Check global wildcard (super admin)
	if perms["*:*"] {
		return true
	}

	return false
}

// IsSuperAdmin checks if the permission map indicates super admin status
func IsSuperAdmin(perms map[string]bool) bool {
	return perms["*:*"]
}

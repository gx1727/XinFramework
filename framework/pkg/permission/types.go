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

// BuildDataScopeSQL builds SQL WHERE clause for data filtering based on DataScope
// This is a utility function that can be used by both UserContext and PermissionService
func BuildDataScopeSQL(ds DataScope, userID uint, orgID int64) (string, []any, error) {
	switch ds.Type {
	case DataScopeAll:
		// No filtering - can see all data
		return "", nil, nil

	case DataScopeSelf:
		return "creator_id = $1", []any{userID}, nil

	case DataScopeCustom:
		if len(ds.OrgIDs) == 0 {
			return "creator_id = $1", []any{userID}, nil
		}
		return "org_id = ANY($1)", []any{ds.OrgIDs}, nil

	case DataScopeDept:
		if orgID == 0 {
			return "creator_id = $1", []any{userID}, nil
		}
		return "org_id = $1", []any{orgID}, nil

	case DataScopeDeptAndBelow:
		if orgID == 0 {
			return "creator_id = $1", []any{userID}, nil
		}
		// Use CTE to find all descendant org IDs
		return `
			org_id = $1
			OR org_id IN (
				WITH RECURSIVE org_tree AS (
					SELECT id FROM organizations WHERE id = $1
					UNION ALL
					SELECT o.id FROM organizations o
					JOIN org_tree ot ON o.parent_id = ot.id
				)
				SELECT id FROM org_tree
			)
		`, []any{orgID}, nil

	default:
		return "creator_id = $1", []any{userID}, nil
	}
}

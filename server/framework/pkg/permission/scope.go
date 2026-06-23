package permission

import "strings"

// ScopeColumns maps data-scope predicates to concrete table columns.
type ScopeColumns struct {
	SelfColumn string
	// SelfUseOrgID switches the "self" value from userID to orgID for organization-like entities.
	SelfUseOrgID bool
	CreatorID    string
	OrgID        string
}

// ScopeFilter is a ready-to-apply SQL predicate and its bind arguments.
type ScopeFilter struct {
	SQL  string
	Args []any
}

var DefaultScopeColumns = ScopeColumns{
	CreatorID: "creator_id",
	OrgID:     "org_id",
}

func (f ScopeFilter) IsEmpty() bool {
	return strings.TrimSpace(f.SQL) == ""
}

func normalizeScopeColumns(columns ScopeColumns) ScopeColumns {
	if columns.SelfColumn == "" {
		columns.SelfColumn = columns.CreatorID
	}
	if columns.CreatorID == "" {
		columns.CreatorID = DefaultScopeColumns.CreatorID
	}
	if columns.OrgID == "" {
		columns.OrgID = DefaultScopeColumns.OrgID
	}
	if columns.SelfColumn == "" {
		columns.SelfColumn = columns.CreatorID
	}
	return columns
}

// BuildDataScopeFilter builds a SQL predicate for the given data-scope rule.
func BuildDataScopeFilter(ds DataScope, userID uint, orgID int64, columns ScopeColumns) (ScopeFilter, error) {
	columns = normalizeScopeColumns(columns)
	selfValue := any(userID)
	if columns.SelfUseOrgID {
		selfValue = orgID
	}

	switch ds.Type {
	case DataScopeAll:
		return ScopeFilter{}, nil

	case DataScopeSelf:
		return ScopeFilter{
			SQL:  columns.SelfColumn + " = $1",
			Args: []any{selfValue},
		}, nil

	case DataScopeCustom:
		if len(ds.OrgIDs) == 0 {
			return ScopeFilter{
				SQL:  columns.SelfColumn + " = $1",
				Args: []any{selfValue},
			}, nil
		}
		return ScopeFilter{
			SQL:  columns.OrgID + " = ANY($1)",
			Args: []any{ds.OrgIDs},
		}, nil

	case DataScopeDept:
		if orgID == 0 {
			return ScopeFilter{
				SQL:  columns.SelfColumn + " = $1",
				Args: []any{selfValue},
			}, nil
		}
		return ScopeFilter{
			SQL:  columns.OrgID + " = $1",
			Args: []any{orgID},
		}, nil

	case DataScopeDeptAndBelow:
		if orgID == 0 {
			return ScopeFilter{
				SQL:  columns.SelfColumn + " = $1",
				Args: []any{selfValue},
			}, nil
		}
		return ScopeFilter{
			SQL: `
			` + columns.OrgID + ` = $1
			OR ` + columns.OrgID + ` IN (
				WITH RECURSIVE org_tree AS (
					SELECT id FROM tenant_organizations WHERE id = $1
					UNION ALL
					SELECT o.id FROM tenant_organizations o
					JOIN org_tree ot ON o.parent_id = ot.id
				)
				SELECT id FROM org_tree
			)
		`,
			Args: []any{orgID},
		}, nil

	default:
		return ScopeFilter{
			SQL:  columns.SelfColumn + " = $1",
			Args: []any{selfValue},
		}, nil
	}
}

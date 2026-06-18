package permission

import (
	"strings"
	"testing"
)

// columnsForTest is a tiny, fully-qualified ScopeColumns instance used
// across the tests below so we can assert against predictable SQL.
var columnsForTest = ScopeColumns{
	SelfColumn: "u.id",
	CreatorID:  "u.creator_id",
	OrgID:      "u.org_id",
}

// TestBuildDataScopeFilter_All: type 1 (全部) → no filter at all.
func TestBuildDataScopeFilter_All(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeAll}, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.IsEmpty() {
		t.Errorf("DataScopeAll must produce empty filter, got SQL=%q", f.SQL)
	}
}

// TestBuildDataScopeFilter_Self: type 5 (本人) → restrict to userID.
func TestBuildDataScopeFilter_Self(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeSelf}, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "u.id = $1" {
		t.Errorf("want SQL=%q, got %q", "u.id = $1", f.SQL)
	}
	if len(f.Args) != 1 || f.Args[0] != uint(42) {
		t.Errorf("want args=[42], got %v", f.Args)
	}
}

// TestBuildDataScopeFilter_Custom_NoOrgs falls back to self when
// the custom scope is configured but no orgs are listed.
func TestBuildDataScopeFilter_Custom_NoOrgs(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeCustom}, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "u.id = $1" {
		t.Errorf("empty OrgIDs must fall back to self filter, got SQL=%q", f.SQL)
	}
}

// TestBuildDataScopeFilter_Custom_WithOrgs: org_id = ANY($1) with the
// org-id slice bound as a single argument.
func TestBuildDataScopeFilter_Custom_WithOrgs(t *testing.T) {
	ds := DataScope{Type: DataScopeCustom, OrgIDs: []int64{1, 2, 3}}
	f, err := BuildDataScopeFilter(ds, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "u.org_id = ANY($1)" {
		t.Errorf("want SQL=%q, got %q", "u.org_id = ANY($1)", f.SQL)
	}
	if len(f.Args) != 1 {
		t.Fatalf("custom filter should bind one array arg, got %v", f.Args)
	}
	got, ok := f.Args[0].([]int64)
	if !ok || len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("org ids mismatch: got %#v", f.Args[0])
	}
}

// TestBuildDataScopeFilter_Dept: with non-zero orgID, restrict to that
// org_id directly.
func TestBuildDataScopeFilter_Dept(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeDept}, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "u.org_id = $1" {
		t.Errorf("want SQL=%q, got %q", "u.org_id = $1", f.SQL)
	}
	if len(f.Args) != 1 || f.Args[0] != int64(7) {
		t.Errorf("want args=[7], got %v", f.Args)
	}
}

// TestBuildDataScopeFilter_Dept_ZeroOrgID: orgID==0 falls back to self
// (no usable dept filter; better to show no rows than everything).
func TestBuildDataScopeFilter_Dept_ZeroOrgID(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeDept}, 42, 0, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "u.id = $1" {
		t.Errorf("orgID=0 must fall back to self, got SQL=%q", f.SQL)
	}
}

// TestBuildDataScopeFilter_DeptAndBelow exercises the recursive CTE.
// We only assert the SQL shape (substring checks) and arg count,
// since hand-asserting the entire CTE is brittle.
func TestBuildDataScopeFilter_DeptAndBelow(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeDeptAndBelow}, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, needle := range []string{"u.org_id = $1", "WITH RECURSIVE", "UNION ALL", "parent_id"} {
		if !strings.Contains(f.SQL, needle) {
			t.Errorf("dept-and-below SQL missing %q\nSQL was:\n%s", needle, f.SQL)
		}
	}
	if len(f.Args) != 1 || f.Args[0] != int64(7) {
		t.Errorf("dept-and-below must bind orgID once, got %v", f.Args)
	}
}

// TestBuildDataScopeFilter_UnknownType: defensive default — unknown
// DataScopeType should not panic; it falls back to self filter.
func TestBuildDataScopeFilter_UnknownType(t *testing.T) {
	f, err := BuildDataScopeFilter(DataScope{Type: DataScopeType(99)}, 42, 7, columnsForTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "u.id = $1" {
		t.Errorf("unknown DataScopeType should fall back to self, got SQL=%q", f.SQL)
	}
}

// TestNormalizeScopeColumns verifies the two-step defaulting rule:
//   1. If SelfColumn is empty, it inherits from CreatorID.
//   2. If CreatorID is empty, it inherits from DefaultScopeColumns.
//   3. If OrgID is empty, it inherits from DefaultScopeColumns.
func TestNormalizeScopeColumns(t *testing.T) {
	// Case 1: caller provides everything.
	got := normalizeScopeColumns(ScopeColumns{
		SelfColumn: "u.id", CreatorID: "u.creator_id", OrgID: "u.org_id",
	})
	if got.SelfColumn != "u.id" || got.CreatorID != "u.creator_id" || got.OrgID != "u.org_id" {
		t.Errorf("passthrough mismatch: %+v", got)
	}

	// Case 2: SelfColumn omitted → fall back to CreatorID.
	got = normalizeScopeColumns(ScopeColumns{
		CreatorID: "u.creator_id", OrgID: "u.org_id",
	})
	if got.SelfColumn != "u.creator_id" {
		t.Errorf("SelfColumn should default to CreatorID, got %q", got.SelfColumn)
	}

	// Case 3: completely empty → all defaults populated.
	got = normalizeScopeColumns(ScopeColumns{})
	if got.CreatorID != DefaultScopeColumns.CreatorID ||
		got.OrgID != DefaultScopeColumns.OrgID {
		t.Errorf("empty columns should be defaulted, got %+v", got)
	}
}
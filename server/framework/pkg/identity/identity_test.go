// Package identity holds cross-domain base types. This file (identity_test.go)
// locks in the field set of each base struct via reflection. Any field addition,
// removal, or JSON-tag rename will fail these tests, forcing the author to
// sync apps/tenant/* and apps/platform/sys_* consumers in lockstep.
//
// 0023.3 prerequisite test base.
package identity

import (
	"reflect"
	"testing"
)

// expectedFields locks each base struct's (field name, JSON tag) list.
// Update this map when adding fields; sync apps/tenant/* MarshalJSON
// (if any) and apps/platform/sys_* downstream consumers at the same time.

var expectedFields = map[string][]fieldSpec{
	"User": {
		{"ID", "id"},
		{"AccountID", "account_id"},
		{"OrgID", "org_id"},
		{"Code", "code"},
		{"RealName", "real_name"},
		{"Nickname", "nickname"},
		{"Avatar", "avatar"},
		{"Status", "status"},
		{"CreatedAt", "created_at"},
		{"UpdatedAt", "updated_at"},
	},
	"Org": {
		{"ID", "id"},
		{"ParentID", "parent_id"},
		{"Code", "code"},
		{"Name", "name"},
		{"Type", "type"},
		{"Description", "description"},
		{"AdminCode", "admin_code"},
		{"Ancestors", "ancestors"},
		{"Sort", "sort"},
		{"Status", "status"},
		{"CreatedAt", "created_at"},
		{"UpdatedAt", "updated_at"},
	},
	"Role": {
		{"ID", "id"},
		{"OrgID", "org_id"},
		{"Code", "code"},
		{"Name", "name"},
		{"Description", "description"},
		{"DataScope", "data_scope"},
		{"IsDefault", "is_default"},
		{"Sort", "sort"},
		{"Status", "status"},
		{"CreatedAt", "created_at"},
		{"UpdatedAt", "updated_at"},
	},
	"Menu": {
		{"ID", "id"},
		{"Code", "code"},
		{"Name", "name"},
		{"Subtitle", "subtitle"},
		{"URL", "url"},
		{"Path", "path"},
		{"Icon", "icon"},
		{"Sort", "sort"},
		{"ParentID", "parent_id"},
		{"Ancestors", "ancestors"},
		{"Visible", "visible"},
		{"Enabled", "enabled"},
		{"CreatedAt", "created_at"},
		{"UpdatedAt", "updated_at"},
	},
	"Permission": {
		{"ID", "id"},
		{"MenuID", "menu_id"},
		{"Code", "code"},
		{"Name", "name"},
		{"Action", "action"},
		{"Description", "description"},
		{"Sort", "sort"},
		{"Status", "status"},
		{"CreatedAt", "created_at"},
		{"UpdatedAt", "updated_at"},
	},
}

type fieldSpec struct {
	Name    string
	JSONTag string
}

func TestUser_FieldSetLocked(t *testing.T) {
	assertFields(t, "User", User{}, expectedFields["User"])
}

func TestOrg_FieldSetLocked(t *testing.T) {
	assertFields(t, "Org", Org{}, expectedFields["Org"])
}

func TestRole_FieldSetLocked(t *testing.T) {
	assertFields(t, "Role", Role{}, expectedFields["Role"])
}

func TestMenu_FieldSetLocked(t *testing.T) {
	assertFields(t, "Menu", Menu{}, expectedFields["Menu"])
}

func TestPermission_FieldSetLocked(t *testing.T) {
	assertFields(t, "Permission", Permission{}, expectedFields["Permission"])
}

// assertFields uses reflection to verify that the struct field count, names,
// and JSON tags match expected exactly. Any mismatch fails the test.
func assertFields(t *testing.T, typeName string, instance interface{}, expected []fieldSpec) {
	t.Helper()
	typ := reflect.TypeOf(instance)
	if typ.NumField() != len(expected) {
		t.Fatalf("%s: NumField = %d, expected %d. After change, sync expectedFields and notify apps/tenant + apps/platform downstream",
			typeName, typ.NumField(), len(expected))
	}
	for i, want := range expected {
		got := typ.Field(i)
		if got.Name != want.Name {
			t.Errorf("%s: field[%d].Name = %q, expected %q", typeName, i, got.Name, want.Name)
		}
		tag := got.Tag.Get("json")
		if tag != want.JSONTag {
			t.Errorf("%s.%s: json tag = %q, expected %q", typeName, got.Name, tag, want.JSONTag)
		}
	}
}

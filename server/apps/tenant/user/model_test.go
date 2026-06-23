package user

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	pkgrbac "gx1727.com/xin/framework/pkg/tenant/auth"
	identity "gx1727.com/xin/framework/pkg/identity"
)

// 0023.3 prerequisite test base: snapshots the current pkgrbac.User
// (which apps/tenant/user.User references via type alias) JSON byte
// output. After 0023.3, User becomes an embed of identity.User plus
// three domain-specific fields plus a custom MarshalJSON. This golden
// JSON is the reference for byte-level compatibility with downstream
// consumers (cms/handler, weixin/service, etc.).
//
// After 0023.3 rename, this test migrates to apps/tenant/user/model_test.go
// and asserts the new User.MarshalJSON() output matches this golden.

// goldenJSON is the snapshot of pkgrbac.User JSON output.
// Field order is fixed: id, tenant_id, account_id, org_id, org_name,
// code, nickname, real_name, avatar, phone, email, status, created_at,
// updated_at. Any field-name / order / type change makes this test fail.
// Note: Go's encoding/json outputs non-ASCII as raw UTF-8 bytes (NOT
// \uXXXX escapes), so the CJK strings appear here as actual UTF-8 bytes.
const goldenJSON = `{"id":1,"tenant_id":2,"account_id":3,"org_id":7,"org_name":"财务部","code":"u001","nickname":"张三","real_name":"张三丰","avatar":"a.png","phone":"13800138000","email":"a@b.com","status":1,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-06-01T00:00:00Z"}`

var fixtureTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixtureUser() pkgrbac.User {
	orgID := uint(7)
	return pkgrbac.User{
		User: identity.User{
			ID:        1,
			AccountID: 3,
			OrgID:     &orgID,
			Code:      "u001",
			RealName:  "张三丰",
			Nickname:  "张三",
			Avatar:    "a.png",
			Status:    1,
			CreatedAt: fixtureTime,
			UpdatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		TenantID: 2,
		Phone:    "13800138000",
		Email:    "a@b.com",
		OrgName:  "财务部",
	}
}

// TestPkgrbacUser_GoldenJSON locks pkgrbac.User JSON bytes. After 0023.3
// rename, apps/tenant/user.User.MarshalJSON() must produce the same bytes.
func TestPkgrbacUser_GoldenJSON(t *testing.T) {
	u := fixtureUser()
	got, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal pkgrbac.User: %v", err)
	}
	if !bytes.Equal(got, []byte(goldenJSON)) {
		t.Fatalf("pkgrbac.User JSON byte-level mismatch\n  got:  %s\n  want: %s", got, goldenJSON)
	}
}

// TestAuthUser_StructShapeLocked locks the 0023.3 struct composition:
// auth.User is an embed of identity.User plus 4 tenant-domain fields.
// JSON output order is verified separately by TestPkgrbacUser_GoldenJSON
// (the custom MarshalJSON pins the legacy 14-field order byte-for-byte).
func TestAuthUser_StructShapeLocked(t *testing.T) {
	typ := reflect.TypeOf(pkgrbac.User{})

	// 1 direct field (the identity.User embed) + 4 extensions = 5.
	if typ.NumField() != 5 {
		t.Fatalf("auth.User NumField = %d, expected 5 (1 embed + 4 extensions)",
			typ.NumField())
	}

	// Field 0: the embed of identity.User.
	if typ.Field(0).Name != "User" {
		t.Errorf("field[0].Name = %q, expected %q (identity.User embed)",
			typ.Field(0).Name, "User")
	}
	if typ.Field(0).Type != reflect.TypeOf(identity.User{}) {
		t.Errorf("field[0].Type = %v, expected identity.User", typ.Field(0).Type)
	}

	// Fields 1-4: tenant-domain extensions in the documented order.
	wantExtensions := []string{"TenantID", "Phone", "Email", "OrgName"}
	for i, want := range wantExtensions {
		got := typ.Field(i + 1).Name
		if got != want {
			t.Errorf("field[%d].Name = %q, expected %q (tenant-domain extension)",
				i+1, got, want)
		}
	}
}

// TestUser_TypeAlias locks that User == pkgrbac.User (type alias).
// After 0023.3 User becomes a distinct struct (embed identity.User +
// domain-specific fields + MarshalJSON). Delete this test then.
func TestUser_TypeAlias(t *testing.T) {
	u := User{}
	if reflect.TypeOf(u) != reflect.TypeOf(pkgrbac.User{}) {
		t.Fatalf("apps/tenant/user.User should still be a type alias of pkgrbac.User. " +
			"0023.3 rename will convert this; remove this test then.")
	}
}

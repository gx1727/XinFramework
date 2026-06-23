package user

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
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
		ID:        1,
		TenantID:  2,
		AccountID: 3,
		OrgID:     &orgID,
		OrgName:   "财务部",
		Code:      "u001",
		Nickname:  "张三",
		RealName:  "张三丰",
		Avatar:    "a.png",
		Phone:     "13800138000",
		Email:     "a@b.com",
		Status:    1,
		CreatedAt: fixtureTime,
		UpdatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
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

// TestPkgrbacUser_FieldOrderLocked locks pkgrbac.User field order via reflection.
// After 0023.3 rename this migrates to apps/tenant/user/ and asserts the
// MarshalJSON output order.
func TestPkgrbacUser_FieldOrderLocked(t *testing.T) {
	expectedOrder := []string{
		"ID", "TenantID", "AccountID", "OrgID", "OrgName",
		"Code", "Nickname", "RealName", "Avatar",
		"Phone", "Email", "Status", "CreatedAt", "UpdatedAt",
	}
	typ := reflect.TypeOf(pkgrbac.User{})
	if typ.NumField() != len(expectedOrder) {
		t.Fatalf("pkgrbac.User NumField = %d, expected %d",
			typ.NumField(), len(expectedOrder))
	}
	for i, want := range expectedOrder {
		got := typ.Field(i).Name
		if got != want {
			t.Errorf("field[%d].Name = %q, expected %q", i, got, want)
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

package xincontext

import (
	"context"
	"testing"

	"gx1727.com/xin/framework/pkg/permission"
)

// TestScopeFilterFrom_NoUserContext 验证 ctx 中没有 UserContext 时
// 返回空 filter（不施加任何过滤），不 panic，不报错。
//
// 这是 ctx-aware helper 的关键约束：
//   - 后台异步任务、CLI 工具、定时任务调用 Repository 时往往不带 UserContext
//   - 这种场景下应"什么都不做"而不是"panic"或"返回错误"
func TestScopeFilterFrom_NoUserContext(t *testing.T) {
	f, err := ScopeFilterFrom(context.Background(), permission.ScopeColumns{
		SelfColumn: "u.id",
		OrgID:      "u.org_id",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.IsEmpty() {
		t.Errorf("empty UserContext must yield empty filter, got SQL=%q", f.SQL)
	}
}

// TestScopeFilterFrom_NilContext 验证 ctx=nil 也不 panic。
func TestScopeFilterFrom_NilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ScopeFilterFrom(nil, ...) must not panic, got: %v", r)
		}
	}()
	//nolint:staticcheck // 故意传入 nil 测试健壮性
	f, _ := ScopeFilterFrom(nil, permission.ScopeColumns{})
	if !f.IsEmpty() {
		t.Errorf("nil ctx must yield empty filter, got SQL=%q", f.SQL)
	}
}

// TestScopeFilterFrom_SelfScope 验证 ctx-aware helper 会从 UserContext
// 取出 DataScope 并委托给 BuildDataScopeFilter 生成 SQL。
func TestScopeFilterFrom_SelfScope(t *testing.T) {
	uc := &UserContext{
		XinContext: &XinContext{UserID: 42},
		DataScope:  permission.DataScope{Type: permission.DataScopeSelf},
	}
	ctx := WithUserContext(context.Background(), uc)

	f, err := ScopeFilterFrom(ctx, permission.ScopeColumns{
		SelfColumn: "u.id",
		OrgID:      "u.org_id",
	})
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

// TestScopeFilterFrom_AllScope 验证 DataScope=ALL 时返回空 filter。
func TestScopeFilterFrom_AllScope(t *testing.T) {
	uc := &UserContext{
		XinContext: &XinContext{UserID: 42},
		DataScope:  permission.DataScope{Type: permission.DataScopeAll},
	}
	ctx := WithUserContext(context.Background(), uc)

	f, err := ScopeFilterFrom(ctx, permission.ScopeColumns{
		SelfColumn: "u.id",
		OrgID:      "u.org_id",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.IsEmpty() {
		t.Errorf("DataScopeAll must yield empty filter, got SQL=%q", f.SQL)
	}
}

// TestScopeFilterFrom_ColumnsApplied 验证传入 columns 的列名会被使用。
func TestScopeFilterFrom_ColumnsApplied(t *testing.T) {
	uc := &UserContext{
		XinContext: &XinContext{UserID: 42},
		OrgID:      7,
		DataScope:  permission.DataScope{Type: permission.DataScopeDept},
	}
	ctx := WithUserContext(context.Background(), uc)

	f, err := ScopeFilterFrom(ctx, permission.ScopeColumns{
		SelfColumn: "x.user_id",
		OrgID:      "x.dept_id",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.SQL != "x.dept_id = $1" {
		t.Errorf("want SQL=%q, got %q", "x.dept_id = $1", f.SQL)
	}
	if len(f.Args) != 1 || f.Args[0] != int64(7) {
		t.Errorf("want args=[7], got %v", f.Args)
	}
}

// TestScopeFilterFrom_LoaderUsed 验证通过 WithUserContextLoader 注册的
// 懒加载器也会被 ctx-aware helper 触发（确保异步场景下也能拿到 scope）。
func TestScopeFilterFrom_LoaderUsed(t *testing.T) {
	called := 0
	loader := func() *UserContext {
		called++
		return &UserContext{
			XinContext: &XinContext{UserID: 99},
			DataScope:  permission.DataScope{Type: permission.DataScopeSelf},
		}
	}
	ctx := WithUserContextLoader(context.Background(), loader)

	f, err := ScopeFilterFrom(ctx, permission.ScopeColumns{SelfColumn: "u.id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Errorf("loader must be invoked exactly once, got %d", called)
	}
	if f.SQL != "u.id = $1" {
		t.Errorf("want SQL=%q, got %q", "u.id = $1", f.SQL)
	}
	if len(f.Args) != 1 || f.Args[0] != uint(99) {
		t.Errorf("want args=[99], got %v", f.Args)
	}
}
package permission

import (
	"context"
	"testing"
	"time"
)

// TestMemoryPermissionCache_GetSetHitMiss 验证基本读写与过期语义。
func TestMemoryPermissionCache_GetSetHitMiss(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryPermissionCache()

	// miss
	got, err := c.GetPermissions(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("miss must return nil map, got %v", got)
	}

	// set + hit
	want := map[string]bool{"user:list": true, "user:create": true}
	if err := c.SetPermissions(ctx, 42, want); err != nil {
		t.Fatalf("SetPermissions: %v", err)
	}
	got, err = c.GetPermissions(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mapsEqual(got, want) {
		t.Errorf("hit must return cached value, got %v want %v", got, want)
	}

	// invalidate
	if err := c.InvalidatePermissions(ctx, 42); err != nil {
		t.Fatalf("InvalidatePermissions: %v", err)
	}
	got, err = c.GetPermissions(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("post-invalidate must be miss, got %v", got)
	}
}

// TestMemoryPermissionCache_DataScope 同上但测数据范围。
func TestMemoryPermissionCache_DataScope(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryPermissionCache()

	ds := &DataScope{Type: DataScopeDept}
	if err := c.SetDataScope(ctx, 7, ds); err != nil {
		t.Fatalf("SetDataScope: %v", err)
	}
	got, err := c.GetDataScope(ctx, 7)
	if err != nil {
		t.Fatalf("GetDataScope: %v", err)
	}
	if got == nil || got.Type != DataScopeDept {
		t.Errorf("hit failed, got %+v", got)
	}

	if err := c.InvalidateDataScope(ctx, 7); err != nil {
		t.Fatalf("InvalidateDataScope: %v", err)
	}
	got, _ = c.GetDataScope(ctx, 7)
	if got != nil {
		t.Errorf("post-invalidate must be miss, got %+v", got)
	}
}

// TestMemoryPermissionCache_TTLExpiry 验证 TTL 到期后过期。
//
// 通过 SetPermTTL(50ms) + SetDataScopeTTL(50ms) 缩短 TTL，
// 然后 sleep 60ms 再读，必须 miss。
func TestMemoryPermissionCache_TTLExpiry(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryPermissionCache()
	c.SetPermTTL(50 * time.Millisecond)
	c.SetDataScopeTTL(50 * time.Millisecond)

	if err := c.SetPermissions(ctx, 1, map[string]bool{"x": true}); err != nil {
		t.Fatalf("set: %v", err)
	}
	time.Sleep(70 * time.Millisecond)
	got, _ := c.GetPermissions(ctx, 1)
	if got != nil {
		t.Errorf("expired entry must miss, got %v", got)
	}

	// DataScope 同理
	ds := &DataScope{Type: DataScopeAll}
	if err := c.SetDataScope(ctx, 1, ds); err != nil {
		t.Fatalf("set ds: %v", err)
	}
	time.Sleep(70 * time.Millisecond)
	gotDS, _ := c.GetDataScope(ctx, 1)
	if gotDS != nil {
		t.Errorf("expired DataScope must miss, got %+v", gotDS)
	}
}

// TestMemoryPermissionCache_DifferentUsersIsolated 验证不同用户的缓存互不干扰。
func TestMemoryPermissionCache_DifferentUsersIsolated(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryPermissionCache()

	if err := c.SetPermissions(ctx, 1, map[string]bool{"a:1": true}); err != nil {
		t.Fatalf("set 1: %v", err)
	}
	if err := c.SetPermissions(ctx, 2, map[string]bool{"b:2": true}); err != nil {
		t.Fatalf("set 2: %v", err)
	}

	got1, _ := c.GetPermissions(ctx, 1)
	got2, _ := c.GetPermissions(ctx, 2)

	if !mapsEqual(got1, map[string]bool{"a:1": true}) {
		t.Errorf("user 1 cache corrupted: %v", got1)
	}
	if !mapsEqual(got2, map[string]bool{"b:2": true}) {
		t.Errorf("user 2 cache corrupted: %v", got2)
	}

	// 失效 user 1 不应影响 user 2
	if err := c.InvalidatePermissions(ctx, 1); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	if got, _ := c.GetPermissions(ctx, 1); got != nil {
		t.Errorf("user 1 should be invalidated, got %v", got)
	}
	if got, _ := c.GetPermissions(ctx, 2); !mapsEqual(got, map[string]bool{"b:2": true}) {
		t.Errorf("user 2 should be intact, got %v", got)
	}
}

// mapsEqual 辅助函数：比较两个 map 是否完全一致（key 集合 + value）。
func mapsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// TestMemoryPermissionCache_SatisfiesInterface 编译期已通过 _ PermissionCache 断言，
// 这里加一个 runtime 双重检查。
func TestMemoryPermissionCache_SatisfiesInterface(t *testing.T) {
	var _ PermissionCache = NewMemoryPermissionCache()
}
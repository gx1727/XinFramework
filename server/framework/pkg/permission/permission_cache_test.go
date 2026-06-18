package permission

import (
	"context"
	"testing"
)

// Tests in this file assume cache.Get() returns nil — i.e. Redis was
// never initialized. The test process never calls cache.Init(), so
// the global cache.Client stays nil. This lets us exercise the
// "cache unavailable" branch of every RedisPermissionCache method.
//
// If you ever run these tests in an environment where another test
// initialised the global cache, you will need to reset it. As of this
// writing, no other permission test does that.

// TestRedisPermissionCache_NilCache_AllMethodsShortCircuit runs every
// method on a fresh RedisPermissionCache and verifies each one returns
// the documented "no cache available" sentinel without error.
func TestRedisPermissionCache_NilCache_AllMethodsShortCircuit(t *testing.T) {
	ctx := context.Background()
	c := NewRedisPermissionCache()

	// Read paths: should return (nil, nil) when cache is unavailable.
	if got, err := c.GetPermissions(ctx, 42); err != nil || got != nil {
		t.Errorf("GetPermissions(nil-cache) = (%v, %v), want (nil, nil)", got, err)
	}
	if got, err := c.GetDataScope(ctx, 42); err != nil || got != nil {
		t.Errorf("GetDataScope(nil-cache) = (%v, %v), want (nil, nil)", got, err)
	}

	// Write paths: should return nil error without doing anything.
	if err := c.SetPermissions(ctx, 42, map[string]bool{"user:list": true}); err != nil {
		t.Errorf("SetPermissions(nil-cache) returned %v, want nil", err)
	}
	if err := c.SetDataScope(ctx, 42, &DataScope{Type: DataScopeAll}); err != nil {
		t.Errorf("SetDataScope(nil-cache) returned %v, want nil", err)
	}
	if err := c.InvalidatePermissions(ctx, 42); err != nil {
		t.Errorf("InvalidatePermissions(nil-cache) returned %v, want nil", err)
	}
	if err := c.InvalidateDataScope(ctx, 42); err != nil {
		t.Errorf("InvalidateDataScope(nil-cache) returned %v, want nil", err)
	}
}

// TestFormatUint is a trivial guard against accidentally switching
// the helper to a non-base-10 formatter (e.g. hex).
func TestFormatUint(t *testing.T) {
	if got := formatUint(0); got != "0" {
		t.Errorf(`formatUint(0) = %q, want "0"`, got)
	}
	if got := formatUint(42); got != "42" {
		t.Errorf(`formatUint(42) = %q, want "42"`, got)
	}
	if got := formatUint(1<<63 - 1); got != "9223372036854775807" {
		t.Errorf(`formatUint(MaxInt64) = %q, want "9223372036854775807"`, got)
	}
}
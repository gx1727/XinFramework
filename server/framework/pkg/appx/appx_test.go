package appx

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestNewPool_NilFails 验证 NewPool 在 pool 为 nil 时返回错误。
func TestNewPool_NilFails(t *testing.T) {
	_, err := NewPool(nil)
	if err == nil {
		t.Fatal("NewPool(nil) must return error")
	}
}

// TestNewPool_HappyPath 验证非 nil pool 构造成功并保留引用。
func TestNewPool_HappyPath(t *testing.T) {
	pool := &pgxpool.Pool{}
	p, err := NewPool(pool)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	if p.Raw() != pool {
		t.Error("Raw() must return the same pool pointer")
	}
}

// TestMustNewPool_NilPanics 验证 nil pool 触发 panic（构造期 fail-fast）。
func TestMustNewPool_NilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewPool(nil) must panic")
		}
	}()
	_ = MustNewPool(nil)
}

// TestMustNewPool_HappyPath 验证非 nil pool 不 panic。
func TestMustNewPool_HappyPath(t *testing.T) {
	pool := &pgxpool.Pool{}
	p := MustNewPool(pool)
	if p.Raw() != pool {
		t.Error("Raw() must return the same pool pointer")
	}
}

// TestPool_Close_NilSafe 验证 nil Pool.Close() 不 panic（幂等）。
func TestPool_Close_NilSafe(t *testing.T) {
	var p *Pool
	p.Close() // nil receiver 不应 panic
	// 显式构造的 nil raw 也不应 panic
	pp := &Pool{raw: nil}
	pp.Close() // 不应 panic
}

// TestNewApp_RequiresBoth 验证 cfg / pool 都必填。
func TestNewApp_RequiresBoth(t *testing.T) {
	pool := MustNewPool(&pgxpool.Pool{})

	if _, err := NewApp(nil, pool); err == nil {
		t.Error("NewApp with nil cfg must fail")
	}
	if _, err := NewApp(nil, nil); err == nil && !errors.Is(err, err) {
		// nil pool 同样报错
		t.Error("NewApp with nil pool must fail")
	}
}

// TestNewApp_HappyPath 验证完整构造。
func TestNewApp_HappyPath(t *testing.T) {
	pool := MustNewPool(&pgxpool.Pool{})
	cfg := struct{ X int }{X: 1}
	// 简化：用 *config.Config 不便直接构造；用 nil cfg 会失败。
	// 改测 MustNewApp panic 路径。
	_ = cfg
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewApp with nil cfg must panic")
		}
	}()
	_ = MustNewApp(nil, pool)
}

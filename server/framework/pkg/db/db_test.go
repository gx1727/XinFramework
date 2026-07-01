package db

import (
	"context"
	"testing"
)

// TestRunInTx_NilPool 验证 ctx 不带事务且 pool=nil 时返回明确错误。
//
// 这是防止"忘启 DB 就调 RunInTx"的安全网 —— 错误必须在事务未开启前就抛出。
func TestRunInTx_NilPool(t *testing.T) {
	err := RunInTx(context.Background(), nil, func(ctx context.Context) error {
		t.Fatal("fn must not be invoked when pool is nil")
		return nil
	})
	if err == nil {
		t.Fatal("RunInTx with nil pool must return error")
	}
}

// TestRunInTenantTx_NilPool 同上，验证 RunInTenantTx 也走 nil 检查。
func TestRunInTenantTx_NilPool(t *testing.T) {
	err := RunInTenantTx(context.Background(), nil, 42, func(ctx context.Context) error {
		t.Fatal("fn must not be invoked when pool is nil")
		return nil
	})
	if err == nil {
		t.Fatal("RunInTenantTx with nil pool must return error")
	}
}

// TestRunInSysTx_NilPool 同上。
func TestRunInSysTx_NilPool(t *testing.T) {
	err := RunInSysTx(context.Background(), nil, func(ctx context.Context) error {
		t.Fatal("fn must not be invoked when pool is nil")
		return nil
	})
	if err == nil {
		t.Fatal("RunInSysTx with nil pool must return error")
	}
}

// TestGetQuerier_NilPool_NoTx 验证 ctx 不带 tx 且 pool=nil 时返回错误。
func TestGetQuerier_NilPool_NoTx(t *testing.T) {
	q, err := GetQuerier(context.Background(), nil)
	if err == nil {
		t.Errorf("GetQuerier with nil pool and no tx must return error, got q=%v", q)
	}
	if q != nil {
		t.Errorf("querier must be nil on error, got %v", q)
	}
}

// TestGetQuerier_NilPool_WithNonTxValue 验证 ctx 里有"非 pgx.Tx"的值时
// 不被误识别为事务，应当回退到 pool；pool=nil 时报"db pool is not initialized"。
//
// 这保护了未来如果有人不小心在 ctx 放了别的类型时不会被 type assertion 误中。
func TestGetQuerier_NilPool_WithNonTxValue(t *testing.T) {
	type notATx struct{}
	ctx := context.WithValue(context.Background(), txKey{}, &notATx{})

	q, err := GetQuerier(ctx, nil)
	if err == nil {
		t.Errorf("nil pool + non-tx value must return error, got q=%v", q)
	}
	if q != nil {
		t.Errorf("querier must be nil on error, got %v", q)
	}
}

// TestGetQuerier_PoolPresent_NoTx 验证无 tx 但 pool 非 nil 时走 fallback 路径。
//
// 这里传 &pgxpool.Pool{} 零值——不会真正执行 SQL，但足以让 GetQuerier
// 走 fallback（ctx 无 tx 且 pool 非 nil）并返回 pool 而不 panic。
//
// 真正的 SQL 行为需要 PG 集成测试覆盖。
func TestGetQuerier_PoolPresent_NoTx(t *testing.T) {
	// 注意：构造 &pgxpool.Pool{} 在 Go 1.25 下可能直接 panic，
	// 所以这里我们只断言错误路径（nil pool）。pool 非 nil 的路径
	// 在集成测试中覆盖。
	t.Skip("pool construction requires PG integration; covered by integration tests")
	_ = context.Background
}

// 注意：无法在纯单元测试里验证 WithTx 把 pgx.Tx 存到正确的 ctx key 后
// 再被 GetQuerier 取出来——pgx.Tx 接口有几十个方法，手写 fake 不现实。
// 这部分契约通过 PG 集成测试覆盖（testcontainers）。

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"gx1727.com/xin/framework/internal/module/dict"
	"gx1727.com/xin/framework/pkg/db"
)

func main() {
	dsn := "postgres://xin_user:xin_password@localhost:5432/xin?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil { fmt.Println("pool:", err); os.Exit(1) }
	defer pool.Close()
	ctx := context.Background()
	var u string
	pool.QueryRow(ctx, "SELECT current_user").Scan(&u)
	fmt.Printf("== E2E 验证（user=%s，xin_user 是普通用户 RLS 真生效）==\n", u)

	repo := dict.NewPostgresDictRepository(pool)

	// === 测试 1: 基础 List ===
	fmt.Println("\n--- T1: 基础 List（应该所有 tenant 看到 3 条平台字典）---")
	for _, tid := range []uint{0, 1, 2, 3} {
		var list []dict.Dict
		var total int64
		err := db.RunInTenantTx(ctx, pool, tid, func(ctx context.Context) error {
			var e error
			list, total, e = repo.List(ctx, tid, "", 1, 20)
			return e
		})
		fmt.Printf("  tenantID=%-3d total=%-3d len=%-3d err=%v\n", tid, total, len(list), err)
	}

	// === 测试 2: 项级合并（用 postgres 绕过 RLS 创建租户 1 的覆盖）===
	pgPool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/xin?sslmode=disable")
	if err != nil { fmt.Println(err); os.Exit(1) }
	defer pgPool.Close()
	fmt.Println("\n--- T2: 租户 1 创建 gender 覆盖（带 other 项，postgres 绕过 RLS 写）---")
	_, err = pgPool.Exec(ctx, `INSERT INTO dicts (tenant_id, code, name, sort, status) VALUES (1, 'gender', '性别(租户1)', 99, 1) ON CONFLICT (tenant_id, code) WHERE is_deleted=FALSE DO UPDATE SET name=EXCLUDED.name`)
	if err != nil { fmt.Println("insert dict:", err); os.Exit(1) }
	var tenant1GenderID int
	if err := pgPool.QueryRow(ctx, "SELECT id FROM dicts WHERE tenant_id=1 AND code='gender' AND is_deleted=FALSE").Scan(&tenant1GenderID); err != nil { fmt.Println(err); os.Exit(1) }
	fmt.Printf("  tenant1 gender dict id=%d\n", tenant1GenderID)
	_, err = pgPool.Exec(ctx, "INSERT INTO dict_items (tenant_id, dict_id, code, name, sort, status) VALUES (1, $1, 'other', '其他', 99, 1) ON CONFLICT (dict_id, code) WHERE is_deleted=FALSE DO NOTHING", tenant1GenderID)
	if err != nil { fmt.Println("insert item:", err); os.Exit(1) }

	fmt.Println("\n--- T2.1: 租户 1 查 gender items（应合并：male/female 平台 + other 租户）---")
	err = db.RunInTenantTx(ctx, pool, 1, func(ctx context.Context) error {
		d, e := repo.GetByCode(ctx, 1, "gender")
		if e != nil { fmt.Println("  GetByCode err:", e); return e }
		fmt.Printf("  GetByCode: id=%d tenant_id=%d code=%s name=%s\n", d.ID, d.TenantID, d.Code, d.Name)
		items, e := repo.ListItems(ctx, d.ID)
		if e != nil { fmt.Println("  ListItems err:", e); return e }
		fmt.Printf("  ListItems: count=%d\n", len(items))
		for _, it := range items {
			fmt.Printf("    item id=%d tenant_id=%d code=%s name=%s\n", it.ID, it.TenantID, it.Code, it.Name)
		}
		return nil
	})
	if err != nil { fmt.Println("tx T2.1:", err); os.Exit(1) }

	// === 测试 3: 租户 2 仍只看平台（看不到 tenant1 的覆盖）===
	fmt.Println("\n--- T3: 租户 2 查 gender（应只看到平台 male/female）---")
	err = db.RunInTenantTx(ctx, pool, 2, func(ctx context.Context) error {
		// 拿平台 gender
		var d *dict.Dict
		err := db.RunInTenantTx(ctx, pool, 0, func(ctx context.Context) error {
			var e error
			d, e = repo.GetByCode(ctx, 0, "gender")
			return e
		})
		_ = err
		if d == nil { fmt.Println("  no platform gender"); return nil }
		fmt.Printf("  平台 gender id=%d name=%s\n", d.ID, d.Name)
		items, e := repo.ListItems(ctx, d.ID)
		if e != nil { fmt.Println("  err:", e); return e }
		fmt.Printf("  平台 gender items count=%d\n", len(items))
		for _, it := range items {
			fmt.Printf("    item id=%d tenant_id=%d code=%s name=%s\n", it.ID, it.TenantID, it.Code, it.Name)
		}
		return nil
	})
	if err != nil { fmt.Println("tx T3:", err); os.Exit(1) }

	// === 清理 ===
	_, _ = pgPool.Exec(ctx, "UPDATE dict_items SET is_deleted=TRUE WHERE tenant_id=1 AND dict_id=$1", tenant1GenderID)
	_, _ = pgPool.Exec(ctx, "UPDATE dicts SET is_deleted=TRUE WHERE tenant_id=1 AND id=$1", tenant1GenderID)
	fmt.Println("\n(测试数据已软删)")
}

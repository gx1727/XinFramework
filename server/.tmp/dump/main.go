package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("XIN_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/xin?sslmode=disable"
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Println("connect err:", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	for _, t := range []string{"dicts", "dict_items"} {
		fmt.Println("=== " + t + " ===")
		rows, err := conn.Query(ctx,
			"SELECT column_name, data_type, is_nullable, column_default "+
				"FROM information_schema.columns "+
				"WHERE table_name=$1 ORDER BY ordinal_position", t)
		if err != nil {
			fmt.Println("query err:", err)
			continue
		}
		for rows.Next() {
			var n, dt, nn string
			var df *string
			if err := rows.Scan(&n, &dt, &nn, &df); err != nil {
				fmt.Println("scan err:", err)
				continue
			}
			d := "<nil>"
			if df != nil {
				d = *df
			}
			fmt.Printf("  %-15s %-25s %-3s %s\n", n, dt, nn, d)
		}
		rows.Close()
	}

	// 也 dump _schema_migrations
	fmt.Println("=== _schema_migrations ===")
	rows, err := conn.Query(ctx, "SELECT version, applied_at FROM _schema_migrations ORDER BY version")
	if err != nil {
		fmt.Println("query err:", err)
	} else {
		for rows.Next() {
			var v string
			var at any
			_ = rows.Scan(&v, &at)
			fmt.Printf("  %s @ %v\n", v, at)
		}
		rows.Close()
	}
}

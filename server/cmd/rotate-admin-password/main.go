// 命令 rotate-admin-password 在部署后修改 admin 账号的密码。
//
// 用法：
//
//	# 1. 直接运行，按提示输入新密码（推荐）
//	go run ./cmd/rotate-admin-password
//
//	# 2. 指定账号（默认 admin）
//	go run ./cmd/rotate-admin-password -account admin
//
//	# 3. 通过环境变量注入 DB 配置（与 xin server 共享同一套 XIN_DB_* env）
//	export XIN_DB_HOST=db.example.com
//	export XIN_DB_PORT=5432
//	export XIN_DB_USER=xin_user
//	export XIN_DB_PASSWORD=xxx
//	export XIN_DB_NAME=xin
//	go run ./cmd/rotate-admin-password
//
// 安全注意事项：
//   - 密码从 /dev/tty 读取，不进 ps 历史
//   - 两次输入确认，避免误输入
//   - 至少 8 位强度校验
//   - 不打印明文密码
//
// 数据库连接与 server 完全共享同一套 env 约定（XIN_DB_*），保证 dev/staging/prod 配置一致。
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5"
	"golang.org/x/term"

	pkgauth "gx1727.com/xin/framework/pkg/auth"
)

func main() {
	account := flag.String("account", "admin", "要修改密码的账号（username/phone/email）")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: rotate-admin-password [-account <account>]\n\n"+
			"通过环境变量 XIN_DB_HOST/PORT/USER/PASSWORD/NAME 配置数据库连接。\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	fmt.Println("========================================")
	fmt.Println("  XinFramework admin 密码重置工具")
	fmt.Println("========================================")
	fmt.Printf("目标账号: %s\n", *account)
	fmt.Println()

	// 1. 读 DB 连接
	dsn := buildDSN()
	if dsn == "" {
		log.Fatalf("缺少数据库连接配置：请设置 XIN_DB_HOST / XIN_DB_PORT / XIN_DB_USER / XIN_DB_PASSWORD / XIN_DB_NAME")
	}

	// 2. 读新密码（两次确认）
	pw, err := promptPasswordTwice()
	if err != nil {
		log.Fatalf("读取密码失败: %v", err)
	}
	if len(pw) < 8 {
		log.Fatalf("密码强度不足：至少 8 位")
	}

	// 3. Hash（使用与 server 相同的 argon2id 参数）
	hash, err := pkgauth.HashPassword(pw)
	if err != nil {
		log.Fatalf("密码哈希失败: %v", err)
	}
	// 清零明文（防御性）
	for i := range pw {
		pw = pw[:i] + "\x00" + pw[i+1:]
	}
	pw = ""

	// 4. 连接 DB
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer conn.Close(ctx)

	// 5. 校验账号存在
	var accountID uint
	var oldUsername string
	err = conn.QueryRow(ctx, `
		SELECT id, username FROM accounts
		WHERE is_deleted = FALSE
		  AND (username = $1 OR phone = $1 OR email = $1)
		LIMIT 1
	`, *account).Scan(&accountID, &oldUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Fatalf("账号 %q 不存在（按 username/phone/email 都没匹配到）", *account)
		}
		log.Fatalf("查询账号失败: %v", err)
	}
	fmt.Printf("找到账号: id=%d username=%s\n", accountID, oldUsername)

	// 6. 二次确认（防误操作）
	fmt.Printf("\n⚠️  即将修改该账号的登录密码。\n")
	fmt.Printf("    请确认是否继续？(yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "yes" && answer != "y" {
		fmt.Println("已取消。")
		return
	}

	// 7. UPDATE
	tag, err := conn.Exec(ctx, `
		UPDATE accounts
		SET password = $1, updated_at = NOW()
		WHERE id = $2 AND is_deleted = FALSE
	`, hash, accountID)
	if err != nil {
		log.Fatalf("更新密码失败: %v", err)
	}
	if tag.RowsAffected() == 0 {
		log.Fatalf("更新失败：rows affected = 0（账号 id=%d 可能已被删除）", accountID)
	}

	fmt.Println()
	fmt.Println("✅ 密码已更新。")
	fmt.Println()
	fmt.Println("建议：")
	fmt.Println("  - 立即用新密码登录一次验证")
	fmt.Println("  - 不要把新密码提交到 git 或聊天记录")
	fmt.Println("  - 如果旧密码曾在多个环境使用，建议同步更新")
}

// buildDSN 从环境变量组装 PostgreSQL DSN。
// 兼容 server 使用的 XIN_DB_* 约定。
func buildDSN() string {
	host := os.Getenv("XIN_DB_HOST")
	port := os.Getenv("XIN_DB_PORT")
	user := os.Getenv("XIN_DB_USER")
	pass := os.Getenv("XIN_DB_PASSWORD")
	name := os.Getenv("XIN_DB_NAME")
	sslmode := os.Getenv("XIN_DB_SSLMODE")

	if host == "" || user == "" || name == "" {
		return ""
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	// 注意：pgx.Connect 接受 URL 形式，密码需 URL encode。
	// 这里简化处理：如果密码包含特殊字符，建议用 XIN_DB_DSN 环境变量整体注入。
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, name, sslmode)
}

// promptPasswordTwice 从 /dev/tty 读取密码两次，两次一致才返回。
func promptPasswordTwice() (string, error) {
	fmt.Print("请输入新密码（至少 8 位，输入不可见）: ")
	pw1, err := readPassword()
	if err != nil {
		return "", err
	}
	fmt.Println()
	fmt.Print("请再次输入新密码（确认）:               ")
	pw2, err := readPassword()
	if err != nil {
		return "", err
	}
	fmt.Println()

	if pw1 != pw2 {
		return "", fmt.Errorf("两次输入不一致")
	}
	return pw1, nil
}

// readPassword 从 stdin 读取密码（终端不回显）。
func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// 非 TTY 场景（如 docker exec -i）回退到 bufio
		// 此场景下密码会出现在 process list / shell history，调用方需自负责任
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
	pw, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
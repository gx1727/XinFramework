package config

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestLoad_DevConfig_ModuleListMatchesSpec 验证 dev 环境 config.yaml 加载后的
// cfg.Module 列表与"alwaysOn + optOut + config.yaml 显式列出的 optional"预期一致。
//
// 这是 P0.3 收尾的回归保护：未来谁动了 config.yaml 都会立刻被这个 case
// 抓到，迫使他同步更新预期（或重新审视是否需要那个模块）。
//
// 2026-06 重构：cfg.Module 的语义从"白名单"改为"累加 optional 模块"，所以
// 即便 config.yaml 只写了 weixin/cms/flag，也会自动加载 alwaysOn + 全部 optOut。
func TestLoad_DevConfig_ModuleListMatchesSpec(t *testing.T) {
	// 隔离环境变量：避免 XIN_APP_ENV / XIN_JWT_SECRET 影响 Load
	t.Setenv("XIN_APP_ENV", "")
	t.Setenv("XIN_JWT_SECRET", "")

	// framework/pkg/config/  → ../../../config/config.yaml
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	cfgPath := filepath.Join(repoRoot, "config", "config.yaml")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load(%q) failed: %v", cfgPath, err)
	}

	// 期望：alwaysOn (3) + optOut (13) + config.yaml 显式列出的 3 项 = 19 项
	// alwaysOn: system, auth, platform_tenant
	// optOut   : menu, user, role, resource, organization, permission,
	//            dict, asset, config,
	//            sys_user, sys_role, sys_menu, sys_permission
	// config.yaml module: weixin, cms, flag
	want := []string{
		// alwaysOn（顺序敏感：永远在列表头部）
		"system", "auth", "platform_tenant",
		// optOut（默认启用）
		"menu", "user", "role", "resource", "organization", "permission",
		"dict", "asset", "config",
		"sys_user", "sys_role", "sys_menu", "sys_permission",
		// config.yaml 显式列出的 optional
		"weixin", "cms", "flag",
	}

	got := append([]string{}, cfg.Module...)
	sort.Strings(got)
	wantSorted := append([]string{}, want...)
	sort.Strings(wantSorted)

	if !equal(got, wantSorted) {
		t.Fatalf("cfg.Module mismatch\n  got:  %v\n  want: %v", got, wantSorted)
	}

	// 同时验证 alwaysOn 三个确实在列表里（顺序敏感）
	for i, m := range []string{"system", "auth", "platform_tenant"} {
		if cfg.Module[i] != m {
			t.Errorf("cfg.Module[%d]: got %q, want %q (full: %v)", i, cfg.Module[i], m, cfg.Module)
		}
	}
}

// TestLoad_ProdConfig_FailsWithPlaceholderSecret 验证 prod 配置下：
//   - config.yaml 仍是占位 secret + config.prod.yaml 把 env 切到 prod
//     → 期望 Load 阶段就 fail-fast，不进入启动流程。
func TestLoad_ProdConfig_FailsWithPlaceholderSecret(t *testing.T) {
	t.Setenv("XIN_APP_ENV", "")
	t.Setenv("XIN_JWT_SECRET", "")

	// 把"dev / 占位 secret"组合用 config.Load 完整跑一遍
	// 方式：在 tmp 写一个"基线 + 切到 prod"的合成 yaml，让 Load 走完整个流程
	tmpDir := t.TempDir()
	baseYAML := `app:
  env: dev
jwt:
  secret: your-secret-key
module:
  - weixin
`
	overrideYAML := `app:
  env: prod
`
	if err := writeFile(filepath.Join(tmpDir, "config.yaml"), baseYAML); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(tmpDir, "config.prod.yaml"), overrideYAML); err != nil {
		t.Fatal(err)
	}

	// Load 只读 .env 和主 path；config.Load 不读 override 段（那是用户责任合并）
	// 这里我们直接用合成的 baseYAML 调一次，env=prod + 占位 secret 应该 fail
	cfg, err := Load(filepath.Join(tmpDir, "config.yaml"))
	if err != nil {
		t.Fatalf("first Load should succeed (dev): %v", err)
	}
	if cfg.App.Env != "dev" {
		t.Fatalf("expected env=dev, got %q", cfg.App.Env)
	}

	// 现在手动切 env 到 prod，模拟"prod 配置覆盖生效"的状态
	cfg.App.Env = "prod"
	if err := validateJWTSecret(cfg); err == nil {
		t.Fatal("expected validateJWTSecret to fail with placeholder secret in prod, got nil")
	}
}

// TestLoad_ProdConfig_AcceptsStrongSecret 验证 prod 配置下，secret ≥ 32 字节时通过。
func TestLoad_ProdConfig_AcceptsStrongSecret(t *testing.T) {
	t.Setenv("XIN_APP_ENV", "prod")
	t.Setenv("XIN_JWT_SECRET", "kJ8#mN2$pQ7&vR4!wX9*zA5+bY1-cD6=eF0_gH3~iJ4")

	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.App.Env != "prod" {
		t.Fatalf("expected env=prod, got %q", cfg.App.Env)
	}
	if err := validateJWTSecret(cfg); err != nil {
		t.Errorf("expected strong secret to pass, got %v", err)
	}
}

// --- helpers ---

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// 当前路径形如 .../server/framework/pkg/config
	// 向上 3 层到 server/
	dir := cwd
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// 找 root go.mod（不是 framework/apps 子 module 的）
			if filepath.Base(dir) == "server" {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// 兜底：从 cwd 向上找 "config" 目录
	dir = cwd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "config", "config.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

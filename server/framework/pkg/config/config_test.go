package config

import (
	"strings"
	"testing"
)

// --- validateModules ---

func TestValidateModules_NoUserConfig_DefaultsToAllOn(t *testing.T) {
	// 用户没写 module:  → alwaysOn + optOut 全部启用
	c := &Config{}
	if err := validateModules(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := append([]string{}, alwaysOnModules...)
	want = append(want, optOutModules...)
	assertModuleList(t, c.Module, want)
}

func TestValidateModules_UserListIsAdditive(t *testing.T) {
	// 2026-06 重构后：module: 不再做白名单过滤，而是"在 alwaysOn + optOut
	// 之上累加 optional 模块"。所以即便用户只写 weixin，optOut 也会全开。
	c := &Config{Module: []string{"weixin"}}
	if err := validateModules(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := append([]string{}, alwaysOnModules...)
	want = append(want, optOutModules...)
	want = append(want, "weixin")
	assertModuleList(t, c.Module, want)
}

func TestValidateModules_AlwaysOnStaysEvenIfNotInUserList(t *testing.T) {
	// 用户列表里没有 system/auth/tenant，也要保留（这是核心模块）
	c := &Config{Module: []string{"weixin", "cms"}}
	if err := validateModules(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	got := c.Module
	for _, must := range []string{"system", "auth", "platform_tenant"} {
		if !contains(got, must) {
			t.Errorf("alwaysOn module %q missing from %v", must, got)
		}
	}
}

func TestValidateModules_DedupAndNormalize(t *testing.T) {
	c := &Config{Module: []string{"  Weixin  ", "weixin", "", "  ", "CMS"}}
	if err := validateModules(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// alwaysOn 必含，weixin / cms 各一次（去重 + 归一为小写）
	got := c.Module
	if !contains(got, "weixin") || !contains(got, "cms") {
		t.Errorf("expected weixin and cms in %v", got)
	}
	if countOccurrences(got, "weixin") != 1 {
		t.Errorf("expected weixin to appear once, got %d times in %v", countOccurrences(got, "weixin"), got)
	}
	if countOccurrences(got, "cms") != 1 {
		t.Errorf("expected cms to appear once, got %d times in %v", countOccurrences(got, "cms"), got)
	}
}

func TestValidateModules_EmptyListTreatedAsNotConfigured(t *testing.T) {
	// []string{} 等价于"没写"，应触发 optOut 默认开启
	c := &Config{Module: []string{}}
	if err := validateModules(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	for _, m := range optOutModules {
		if !contains(c.Module, m) {
			t.Errorf("optOut module %q should be enabled by default, got %v", m, c.Module)
		}
	}
}

// --- validateJWTSecret ---

func TestValidateJWTSecret_DevAllowsAnything(t *testing.T) {
	// dev 环境不校验
	cases := []string{"", "your-secret-key", "short", "x"}
	for _, secret := range cases {
		c := &Config{
			App: AppConfig{Env: "dev"},
			JWT: JWTConfig{Secret: secret},
		}
		if err := validateJWTSecret(c); err != nil {
			t.Errorf("dev env should accept secret=%q, got %v", secret, err)
		}
	}
}

func TestValidateJWTSecret_ProdRejectsEmpty(t *testing.T) {
	c := &Config{
		App: AppConfig{Env: "prod"},
		JWT: JWTConfig{Secret: ""},
	}
	err := validateJWTSecret(c)
	if err == nil {
		t.Fatal("expected error for empty secret in prod, got nil")
	}
	if !strings.Contains(err.Error(), "FATAL") {
		t.Errorf("error should start with FATAL, got: %v", err)
	}
}

func TestValidateJWTSecret_ProdRejectsPlaceholders(t *testing.T) {
	placeholders := []string{"your-secret-key", "changeme", "please-change-me", "secret", "12345678"}
	for _, p := range placeholders {
		c := &Config{
			App: AppConfig{Env: "prod"},
			JWT: JWTConfig{Secret: p},
		}
		if err := validateJWTSecret(c); err == nil {
			t.Errorf("expected error for placeholder %q in prod, got nil", p)
		}
	}
}

func TestValidateJWTSecret_ProdRejectsShortSecret(t *testing.T) {
	// 31 字节应被拒，32 字节应通过
	short := strings.Repeat("a", 31)
	c := &Config{
		App: AppConfig{Env: "prod"},
		JWT: JWTConfig{Secret: short},
	}
	if err := validateJWTSecret(c); err == nil {
		t.Errorf("expected error for 31-byte secret, got nil")
	}

	ok := strings.Repeat("a", 32)
	c2 := &Config{
		App: AppConfig{Env: "prod"},
		JWT: JWTConfig{Secret: ok},
	}
	if err := validateJWTSecret(c2); err != nil {
		t.Errorf("expected nil for 32-byte secret, got %v", err)
	}
}

func TestValidateJWTSecret_ProdAcceptsLongRandomSecret(t *testing.T) {
	c := &Config{
		App: AppConfig{Env: "prod"},
		JWT: JWTConfig{Secret: "kJ8#mN2$pQ7&vR4!wX9*zA5+bY1-cD6=eF0_gH3~iJ4"},
	}
	if err := validateJWTSecret(c); err != nil {
		t.Errorf("expected nil for strong secret, got %v", err)
	}
}

// --- helpers ---

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func countOccurrences(haystack []string, needle string) int {
	n := 0
	for _, s := range haystack {
		if s == needle {
			n++
		}
	}
	return n
}

func assertModuleList(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("module list length: got %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	// 顺序敏感：alwaysOn 先，optOut 后，用户列表原顺序
	for i, m := range want {
		if got[i] != m {
			t.Errorf("module[%d]: got %q, want %q (full: %v)", i, got[i], m, got)
		}
	}
}

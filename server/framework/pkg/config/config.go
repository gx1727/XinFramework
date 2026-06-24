package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type StorageConfig struct {
	Provider string `yaml:"provider"` // local | cos

	// Local storage
	LocalDir     string `yaml:"local_dir"`
	LocalBaseURL string `yaml:"local_base_url"`

	// COS storage
	CosURL       string `yaml:"cos_url"` // https://<bucket>.cos.<region>.myqcloud.com
	CosSecretID  string `yaml:"cos_secret_id"`
	CosSecretKey string `yaml:"cos_secret_key"`
	CosBaseURL   string `yaml:"cos_base_url"` // https://img.gx1727.com
}

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Storage  StorageConfig  `yaml:"storage"`
	Log      LogConfig      `yaml:"log"`
	Module   []string       `yaml:"module"`
	Apps     []string       `yaml:"apps"`
	CORS     CORSConfig     `yaml:"cors"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	DBName             string `yaml:"dbname"`
	SSLMode            string `yaml:"sslmode"`
	MaxOpenConns       int    `yaml:"max_open_conns"`
	MaxIdleConns       int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSec int    `yaml:"conn_max_lifetime_sec"`
	ConnMaxIdleTimeSec int    `yaml:"conn_max_idle_time_sec"`
}

type RedisConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	Enabled        bool   `yaml:"enabled"`
	Required       bool   `yaml:"required"`
	PoolSize       int    `yaml:"pool_size"`
	MinIdleConns   int    `yaml:"min_idle_conns"`
	PoolTimeoutSec int    `yaml:"pool_timeout_sec"`
	IdleTimeoutSec int    `yaml:"idle_timeout_sec"`
	MaxConnAgeSec  int    `yaml:"max_conn_age_sec"`
}

type JWTConfig struct {
	Secret        string `yaml:"secret"`
	Expire        int    `yaml:"expire"`
	RefreshExpire int    `yaml:"refresh_expire"`
}

type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     string   `yaml:"allow_methods"`
	AllowHeaders     string   `yaml:"allow_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

func (c *CORSConfig) IsEnabled() bool {
	return c.Enabled
}

type LogConfig struct {
	Dir   string `yaml:"dir"`
	Level string `yaml:"level"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func defaults() *Config {
	return &Config{
		App: AppConfig{
			Name: "xin",
			Env:  "dev",
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:               "localhost",
			Port:               5432,
			User:               "xin_user",
			Password:           "xin_password",
			DBName:             "xin",
			SSLMode:            "disable",
			MaxOpenConns:       100,
			MaxIdleConns:       20,
			ConnMaxLifetimeSec: 300,
			ConnMaxIdleTimeSec: 60,
		},
		Redis: RedisConfig{
			Host:           "127.0.0.1",
			Port:           6379,
			PoolSize:       10,
			MinIdleConns:   5,
			PoolTimeoutSec: 4,
			IdleTimeoutSec: 300,
		},
		JWT: JWTConfig{
			Secret:        "",
			Expire:        3600,
			RefreshExpire: 86400,
		},
		Storage: StorageConfig{
			Provider:     "local",
			LocalDir:     "./uploads",
			LocalBaseURL: "/uploads",
		},
		CORS: CORSConfig{
			Enabled:          true,
			AllowOrigins:     []string{"*"},
			AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
			AllowHeaders:     "Content-Type,Authorization,X-Requested-With,X-Request-ID,X-Tenant-ID",
			AllowCredentials: false,
			MaxAge:           86400,
		},
		Log: LogConfig{
			Dir:   "logs",
			Level: "info",
		},
	}
}

func Load(path string) (*Config, error) {
	if err := loadEnv(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env failed: %w", err)
	}

	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else if len(data) > 0 {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	overrideWithEnv(cfg)
	if err := validateModules(cfg); err != nil {
		return nil, err
	}
	if err := validateJWTSecret(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}

	return scanner.Err()
}

func overrideWithEnv(c *Config) {
	envStr("XIN_APP_NAME", &c.App.Name)
	envStr("XIN_APP_ENV", &c.App.Env)
	envStr("XIN_APP_HOST", &c.App.Host)
	envInt("XIN_APP_PORT", &c.App.Port)

	envStr("XIN_DB_HOST", &c.Database.Host)
	envInt("XIN_DB_PORT", &c.Database.Port)
	envStr("XIN_DB_USER", &c.Database.User)
	envStr("XIN_DB_PASSWORD", &c.Database.Password)
	envStr("XIN_DB_NAME", &c.Database.DBName)
	envStr("XIN_DB_SSLMODE", &c.Database.SSLMode)
	envInt("XIN_DB_MAX_OPEN_CONNS", &c.Database.MaxOpenConns)
	envInt("XIN_DB_MAX_IDLE_CONNS", &c.Database.MaxIdleConns)
	envInt("XIN_DB_CONN_MAX_LIFETIME_SEC", &c.Database.ConnMaxLifetimeSec)
	envInt("XIN_DB_CONN_MAX_IDLE_TIME_SEC", &c.Database.ConnMaxIdleTimeSec)

	envStr("XIN_REDIS_HOST", &c.Redis.Host)
	envInt("XIN_REDIS_PORT", &c.Redis.Port)
	envStr("XIN_REDIS_PASSWORD", &c.Redis.Password)
	envInt("XIN_REDIS_DB", &c.Redis.DB)
	envBool("XIN_REDIS_ENABLED", &c.Redis.Enabled)
	envBool("XIN_REDIS_REQUIRED", &c.Redis.Required)
	envInt("XIN_REDIS_POOL_SIZE", &c.Redis.PoolSize)
	envInt("XIN_REDIS_MIN_IDLE_CONNS", &c.Redis.MinIdleConns)
	envInt("XIN_REDIS_POOL_TIMEOUT_SEC", &c.Redis.PoolTimeoutSec)
	envInt("XIN_REDIS_IDLE_TIMEOUT_SEC", &c.Redis.IdleTimeoutSec)
	envInt("XIN_REDIS_MAX_CONN_AGE_SEC", &c.Redis.MaxConnAgeSec)

	envStr("XIN_JWT_SECRET", &c.JWT.Secret)
	envInt("XIN_JWT_EXPIRE", &c.JWT.Expire)
	envInt("XIN_JWT_REFRESH_EXPIRE", &c.JWT.RefreshExpire)

	envStr("XIN_STORAGE_PROVIDER", &c.Storage.Provider)
	envStr("XIN_STORAGE_LOCAL_DIR", &c.Storage.LocalDir)
	envStr("XIN_STORAGE_LOCAL_BASE_URL", &c.Storage.LocalBaseURL)
	envStr("XIN_STORAGE_COS_URL", &c.Storage.CosURL)
	envStr("XIN_STORAGE_COS_SECRET_ID", &c.Storage.CosSecretID)
	envStr("XIN_STORAGE_COS_SECRET_KEY", &c.Storage.CosSecretKey)
	envStr("XIN_STORAGE_COS_BASE_URL", &c.Storage.CosBaseURL)

	envBool("XIN_CORS_ENABLED", &c.CORS.Enabled)
	if origins := os.Getenv("XIN_CORS_ALLOW_ORIGINS"); origins != "" {
		c.CORS.AllowOrigins = strings.Split(origins, ",")
	}
	envStr("XIN_CORS_ALLOW_METHODS", &c.CORS.AllowMethods)
	envStr("XIN_CORS_ALLOW_HEADERS", &c.CORS.AllowHeaders)
	envBool("XIN_CORS_ALLOW_CREDENTIALS", &c.CORS.AllowCredentials)
	envInt("XIN_CORS_MAX_AGE", &c.CORS.MaxAge)

	envStr("XIN_LOG_DIR", &c.Log.Dir)
	envStr("XIN_LOG_LEVEL", &c.Log.Level)
	envCSV("XIN_MODULE", &c.Module)
}

func envStr(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

func envInt(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}

func envBool(key string, target *bool) {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			*target = b
		}
	}
}

func envCSV(key string, target *[]string) {
	if v := os.Getenv(key); v != "" {
		raw := strings.Split(v, ",")
		out := make([]string, 0, len(raw))
		for _, s := range raw {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				out = append(out, s)
			}
		}
		*target = out
	}
}

// alwaysOnModules 启动必需，配置无法禁用。
//
// 这些模块要么承载进程级基础设施（system 提供 health/cache stats），
// 要么被 auth 中间件或框架其它部分隐式依赖（auth / platform_tenant）。
// 关闭它们会导致框架不可用，因此不允许在 module: 里"删一行"就关掉。
var alwaysOnModules = []string{
	"system",
	"auth",
	"platform_tenant",
}

// optOutModules 默认启用，但用户在 module: 中显式列出模块时，改为白名单语义。
//
// 这些是"基础业务套件"——绝大多数部署都会用到，框架帮用户开了；
// 一旦用户写 module:，就视为"我只想要这些"，optOut 全部退出。
// 这样 developing.md 承诺的"module: 删一行就能关掉"才真的成立。
var optOutModules = []string{
	"menu",
	"user",
	"role",
	"resource",
	"organization",
	"dict",
	"asset",
	"permission",
}

func validateModules(c *Config) error {
	seen := map[string]struct{}{}
	var merged []string
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		merged = append(merged, name)
	}

	// 1. alwaysOn 永远在
	for _, m := range alwaysOnModules {
		add(m)
	}

	// 2. 归一用户列表
	userList := make([]string, 0, len(c.Module))
	for _, raw := range c.Module {
		m := strings.ToLower(strings.TrimSpace(raw))
		if m == "" {
			continue
		}
		userList = append(userList, m)
	}

	if len(userList) == 0 {
		// 用户没写 module: → optOut 全部默认启用
		for _, m := range optOutModules {
			add(m)
		}
	} else {
		// 用户写了 module: → 白名单语义：alwaysOn + 用户列表
		for _, m := range userList {
			add(m)
		}
	}

	c.Module = merged
	return nil
}

// jwtSecretPlaceholders 启动时拒绝的占位符。
// 在 prod 环境下 secret 等于其中任一值时，配置加载直接失败。
var jwtSecretPlaceholders = map[string]struct{}{
	"":                 {},
	"your-secret-key":  {},
	"changeme":         {},
	"please-change-me": {},
	"secret":           {},
	"12345678":         {},
}

// validateJWTSecret 在 prod 环境强制要求 jwt.secret 配置正确。
//
// 校验项：
//  1. 不能为空
//  2. 不能是常见占位符（防开发者忘记改）
//  3. 长度必须 ≥ 32 字节（保证 HS256 签名强度，避免短 secret 被爆破）
//
// 失败直接返回 error，由 config.Load 的调用方（main.go）log.Fatalf 退出。
// 之所以在 Load 阶段 fail-fast 而不是启动后 panic，是为了让 CI / docker
// entrypoint 在早期就拿到明确错误信息。
func validateJWTSecret(c *Config) error {
	if c.App.Env != "prod" {
		return nil
	}
	if _, isPlaceholder := jwtSecretPlaceholders[c.JWT.Secret]; isPlaceholder {
		return fmt.Errorf(
			"FATAL: jwt.secret 未配置或仍是占位符 (got=%q); prod 环境要求配置一个 ≥32 字节的随机串",
			c.JWT.Secret,
		)
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf(
			"FATAL: jwt.secret 长度 %d 字节 < 32; prod 环境要求 ≥32 字节随机串以保证 HS256 签名强度",
			len(c.JWT.Secret),
		)
	}
	return nil
}

func (c *Config) ModuleEnabled(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, d := range c.Module {
		if d == name {
			return true
		}
	}
	return false
}

func (c *Config) AppEnabled(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, a := range c.Apps {
		if strings.ToLower(strings.TrimSpace(a)) == name {
			return true
		}
	}
	return false
}

var moduleBaseDir = "config"

func LoadModule(name string, target any) error {
	path := filepath.Join(moduleBaseDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err == nil {
		if len(data) > 0 {
			if err := yaml.Unmarshal(data, target); err != nil {
				return fmt.Errorf("parse module config %s: %w", name, err)
			}
		}
		overrideModuleEnv(name, target)
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("read module config %s: %w", name, err)
	}

	mainPath := filepath.Join(moduleBaseDir, "config.yaml")
	mainData, err := os.ReadFile(mainPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read main config for module %s: %w", name, err)
	}
	if len(mainData) == 0 {
		return nil
	}

	var raw map[string]any
	if err := yaml.Unmarshal(mainData, &raw); err != nil {
		return fmt.Errorf("parse main config for module %s: %w", name, err)
	}

	section, ok := raw[name]
	if !ok {
		return nil
	}

	sectionData, err := yaml.Marshal(section)
	if err != nil {
		return fmt.Errorf("marshal module section %s: %w", name, err)
	}
	if len(sectionData) == 0 {
		return nil
	}
	if err := yaml.Unmarshal(sectionData, target); err != nil {
		return fmt.Errorf("parse module config %s from main: %w", name, err)
	}
	overrideModuleEnv(name, target)
	return nil
}

func overrideModuleEnv(module string, target any) {
	prefix := "XIN_" + strings.ToUpper(module) + "_"
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		yamlKey := field.Tag.Get("yaml")
		if idx := strings.Index(yamlKey, ","); idx != -1 {
			yamlKey = yamlKey[:idx]
		}
		if yamlKey == "" || yamlKey == "-" {
			continue
		}
		envKey := prefix + strings.ToUpper(yamlKey)
		envVal := os.Getenv(envKey)
		if envVal == "" {
			continue
		}
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.SetString(envVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if n, err := strconv.ParseInt(envVal, 10, 64); err == nil {
				f.SetInt(n)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if n, err := strconv.ParseUint(envVal, 10, 64); err == nil {
				f.SetUint(n)
			}
		case reflect.Bool:
			if b, err := strconv.ParseBool(envVal); err == nil {
				f.SetBool(b)
			}
		case reflect.Float32, reflect.Float64:
			if f2, err := strconv.ParseFloat(envVal, 64); err == nil {
				f.SetFloat(f2)
			}
		}
	}
}

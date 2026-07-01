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
	App             AppConfig             `yaml:"app"`
	Database        DatabaseConfig        `yaml:"database"`
	Redis           RedisConfig           `yaml:"redis"`
	JWT             JWTConfig             `yaml:"jwt"`
	Storage         StorageConfig         `yaml:"storage"`
	Log             LogConfig             `yaml:"log"`
	Module          []string              `yaml:"module"`
	Apps            []string              `yaml:"apps"`
	CORS            CORSConfig            `yaml:"cors"`
	PermissionCache PermissionCacheConfig `yaml:"permission_cache"`
	LoginSecurity   LoginSecurityConfig   `yaml:"login_security"`
	Task            TaskConfig            `yaml:"task"`
}

// PermissionCacheConfig 控制权限 / 数据范围缓存的行为。
//
// 装配路径：boot.Init 根据 cfg.Redis.Enabled 决定使用 RedisPermissionCache
// 还是 MemoryPermissionCache。本配置仅控制缓存参数（TTL / key 前缀），
// 缓存类型本身由 Redis 是否可用决定。
type PermissionCacheConfig struct {
	// PermTTLSeconds 权限码缓存 TTL（秒）。默认 900 (15 分钟)。
	PermTTLSeconds int `yaml:"perm_ttl_seconds"`
	// DataScopeTTLSeconds 数据范围缓存 TTL（秒）。默认 1800 (30 分钟)。
	DataScopeTTLSeconds int `yaml:"data_scope_ttl_seconds"`
	// KeyPrefix Redis key 前缀，默认 "user:"。修改后可与同 Redis 上的其他服务隔离。
	KeyPrefix string `yaml:"key_prefix"`
}

// LoginSecurityConfig 控制账号锁定与异地告警的行为。
//
// 详见 framework/pkg/login_security 包与 doc/login-security.md。
type LoginSecurityConfig struct {
	Enabled bool `yaml:"enabled"`

	// 账号维度锁定
	MaxFailedAttempts int `yaml:"max_failed_attempts"` // 滑动窗口内最大失败次数，默认 5
	LockDurationMin   int `yaml:"lock_duration_min"`   // 锁定时长（分钟），默认 30
	FailureWindowMin  int `yaml:"failure_window_min"`  // 滑动窗口（分钟），默认 10

	// IP 维度封锁（跨账号防爆破）
	IPFailureThreshold int `yaml:"ip_failure_threshold"`  // 默认 20
	IPFailureWindowMin int `yaml:"ip_failure_window_min"` // 默认 5

	// 异地告警
	AnomalyEnabled      bool `yaml:"anomaly_enabled"`
	AnomalyHistoryLimit int  `yaml:"anomaly_history_limit"` // 比对最近 N 次，默认 5
	AnomalyDeviceMatch  bool `yaml:"anomaly_device_match"`  // device_id 是否参与判定，默认 false
	AnomalyNotifyInSite bool `yaml:"anomaly_notify_in_site"`
	AnomalyNotifyEmail  bool `yaml:"anomaly_notify_email"`
	AnomalyNotifySMS    bool `yaml:"anomaly_notify_sms"`

	// 锁定通知
	LockNotifyInSite bool `yaml:"lock_notify_in_site"`
	LockNotifyEmail  bool `yaml:"lock_notify_email"`
	LockNotifySMS    bool `yaml:"lock_notify_sms"`
}

// TaskConfig 控制长时任务系统的运行参数。
//
// 详见 framework/pkg/task 包与 doc/task-design.md。
type TaskConfig struct {
	WorkerCount          int               `yaml:"worker_count"`           // 进程内 worker goroutine 数（默认 4）
	PollIntervalMs       int               `yaml:"poll_interval_ms"`       // 轮询间隔（默认 1000）
	HeartbeatIntervalSec int               `yaml:"heartbeat_interval_sec"` // 心跳间隔（默认 30）
	HeartbeatTimeoutSec  int               `yaml:"heartbeat_timeout_sec"`  // 心跳超时视为僵死（默认 90）
	ReclaimIntervalSec   int               `yaml:"reclaim_interval_sec"`   // 僵死回收周期（默认 60）
	DefaultMaxAttempts   int               `yaml:"default_max_attempts"`   // 默认重试次数（默认 3）
	DefaultTimeoutSec    int               `yaml:"default_timeout_sec"`    // 默认单任务超时（默认 300）
	RetryStrategy        string            `yaml:"retry_strategy"`         // exponential/linear/fixed（默认 exponential）
	Cleanup              TaskCleanupConfig `yaml:"cleanup"`
	Cron                 TaskCronConfig    `yaml:"cron"`
}

// IsEnabled cron 调度器总开关。
//
// false 时 module.go 完全跳过 CronScheduler 启动 + 管理 API 注册，
// 但 PGCronStore 仍可被其他代码调用（如 TriggerNow）。
func (c TaskCronConfig) IsEnabled() bool { return c.Enabled }

// TaskCronConfig 控制 cron 周期性调度器的行为。
//
// 详见 framework/pkg/task/cron.go 与 doc/task-cron.md。
type TaskCronConfig struct {
	Enabled          bool `yaml:"enabled"`           // 是否启用 cron 调度器
	ScanIntervalSec  int  `yaml:"scan_interval_sec"` // scanner 周期（默认 60）
	RegisterDefaults bool `yaml:"register_defaults"` // 启动期是否注册框架默认 cron job
}

// TaskCleanupConfig 控制后台任务清理历史周期。
type TaskCleanupConfig struct {
	SucceededKeepDays int `yaml:"succeeded_keep_days"` // succeeded 保留天数（默认 7）
	FailedKeepDays    int `yaml:"failed_keep_days"`    // failed 保留天数（默认 30）
	DeadKeepDays      int `yaml:"dead_keep_days"`      // dead 保留天数（默认 90）
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
		PermissionCache: PermissionCacheConfig{
			PermTTLSeconds:      900,  // 15 min
			DataScopeTTLSeconds: 1800, // 30 min
			KeyPrefix:           "user:",
		},
		LoginSecurity: LoginSecurityConfig{
			Enabled:             true,
			MaxFailedAttempts:   5,
			LockDurationMin:     30,
			FailureWindowMin:    10,
			IPFailureThreshold:  20,
			IPFailureWindowMin:  5,
			AnomalyEnabled:      true,
			AnomalyHistoryLimit: 5,
			AnomalyDeviceMatch:  false,
			AnomalyNotifyInSite: true,
			AnomalyNotifyEmail:  true,
			AnomalyNotifySMS:    false,
			LockNotifyInSite:    true,
			LockNotifyEmail:     true,
			LockNotifySMS:       true,
		},
		Task: TaskConfig{
			WorkerCount:          4,
			PollIntervalMs:       1000,
			HeartbeatIntervalSec: 30,
			HeartbeatTimeoutSec:  90,
			ReclaimIntervalSec:   60,
			DefaultMaxAttempts:   3,
			DefaultTimeoutSec:    300,
			RetryStrategy:        "exponential",
			Cleanup: TaskCleanupConfig{
				SucceededKeepDays: 7,
				FailedKeepDays:    30,
				DeadKeepDays:      90,
			},
			Cron: TaskCronConfig{
				Enabled:          true,
				ScanIntervalSec:  60,
				RegisterDefaults: true,
			},
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

	envInt("XIN_PERMISSION_CACHE_PERM_TTL_SECONDS", &c.PermissionCache.PermTTLSeconds)
	envInt("XIN_PERMISSION_CACHE_DATA_SCOPE_TTL_SECONDS", &c.PermissionCache.DataScopeTTLSeconds)
	envStr("XIN_PERMISSION_CACHE_KEY_PREFIX", &c.PermissionCache.KeyPrefix)

	envBool("XIN_LOGIN_SECURITY_ENABLED", &c.LoginSecurity.Enabled)
	envInt("XIN_LOGIN_SECURITY_MAX_FAILED_ATTEMPTS", &c.LoginSecurity.MaxFailedAttempts)
	envInt("XIN_LOGIN_SECURITY_LOCK_DURATION_MIN", &c.LoginSecurity.LockDurationMin)
	envInt("XIN_LOGIN_SECURITY_FAILURE_WINDOW_MIN", &c.LoginSecurity.FailureWindowMin)
	envInt("XIN_LOGIN_SECURITY_IP_FAILURE_THRESHOLD", &c.LoginSecurity.IPFailureThreshold)
	envInt("XIN_LOGIN_SECURITY_IP_FAILURE_WINDOW_MIN", &c.LoginSecurity.IPFailureWindowMin)
	envBool("XIN_LOGIN_SECURITY_ANOMALY_ENABLED", &c.LoginSecurity.AnomalyEnabled)
	envInt("XIN_LOGIN_SECURITY_ANOMALY_HISTORY_LIMIT", &c.LoginSecurity.AnomalyHistoryLimit)
	envBool("XIN_LOGIN_SECURITY_ANOMALY_DEVICE_MATCH", &c.LoginSecurity.AnomalyDeviceMatch)
	envBool("XIN_LOGIN_SECURITY_ANOMALY_NOTIFY_IN_SITE", &c.LoginSecurity.AnomalyNotifyInSite)
	envBool("XIN_LOGIN_SECURITY_ANOMALY_NOTIFY_EMAIL", &c.LoginSecurity.AnomalyNotifyEmail)
	envBool("XIN_LOGIN_SECURITY_ANOMALY_NOTIFY_SMS", &c.LoginSecurity.AnomalyNotifySMS)
	envBool("XIN_LOGIN_SECURITY_LOCK_NOTIFY_IN_SITE", &c.LoginSecurity.LockNotifyInSite)
	envBool("XIN_LOGIN_SECURITY_LOCK_NOTIFY_EMAIL", &c.LoginSecurity.LockNotifyEmail)
	envBool("XIN_LOGIN_SECURITY_LOCK_NOTIFY_SMS", &c.LoginSecurity.LockNotifySMS)

	envInt("XIN_TASK_WORKER_COUNT", &c.Task.WorkerCount)
	envInt("XIN_TASK_POLL_INTERVAL_MS", &c.Task.PollIntervalMs)
	envInt("XIN_TASK_HEARTBEAT_INTERVAL_SEC", &c.Task.HeartbeatIntervalSec)
	envInt("XIN_TASK_HEARTBEAT_TIMEOUT_SEC", &c.Task.HeartbeatTimeoutSec)
	envInt("XIN_TASK_RECLAIM_INTERVAL_SEC", &c.Task.ReclaimIntervalSec)
	envInt("XIN_TASK_DEFAULT_MAX_ATTEMPTS", &c.Task.DefaultMaxAttempts)
	envInt("XIN_TASK_DEFAULT_TIMEOUT_SEC", &c.Task.DefaultTimeoutSec)
	envStr("XIN_TASK_RETRY_STRATEGY", &c.Task.RetryStrategy)
	envInt("XIN_TASK_CLEANUP_SUCCEEDED_KEEP_DAYS", &c.Task.Cleanup.SucceededKeepDays)
	envInt("XIN_TASK_CLEANUP_FAILED_KEEP_DAYS", &c.Task.Cleanup.FailedKeepDays)
	envInt("XIN_TASK_CLEANUP_DEAD_KEEP_DAYS", &c.Task.Cleanup.DeadKeepDays)

	envBool("XIN_TASK_CRON_ENABLED", &c.Task.Cron.Enabled)
	envInt("XIN_TASK_CRON_SCAN_INTERVAL_SEC", &c.Task.Cron.ScanIntervalSec)
	envBool("XIN_TASK_CRON_REGISTER_DEFAULTS", &c.Task.Cron.RegisterDefaults)
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
// 要么被 auth 中间件或框架其它部分隐式依赖（auth / sys_tenant）。
// 关闭它们会导致框架不可用，因此不允许在 module: 里"删一行"就关掉。
var alwaysOnModules = []string{
	"system",
	"auth",
	"sys_tenant",
}

// optOutModules 默认启用。框架默认加载，无需在 cfg.Module 中声明。
//
// 这些是"框架默认能力"——绝大多数部署都会用到，由框架统一开启。
// 若某个部署确实需要关闭某个 optOut 模块，请直接编辑本列表（不要通过 cfg.Module
// 间接开关——cfg.Module 现在的语义是"累加 optional 模块"，不再做白名单过滤）。
//
// 三档分类（详见 doc/architecture.md §3.3）：
//   - alwaysOn  3  : system / auth / sys_tenant        （永远启用，不可关）
//   - optOut   13  : RBAC + 字典 + 资产 + 配置 + 平台管理  （默认全开）
//   - optional  3  : weixin / cms / flag  （默认关，纯业务/集成）
var optOutModules = []string{
	// 租户域 RBAC 套件
	"menu",
	"user",
	"role",
	"resource",
	"organization",
	"permission",
	// 租户域基础设施
	"dict",
	"asset",
	"config",
	// 平台管理域
	"sys_user",
	"sys_role",
	"sys_menu",
	"sys_permission",
	// 长时任务
	"task",
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

	// 2. optOut 默认全开（框架能力，不需要用户显式声明）
	for _, m := range optOutModules {
		add(m)
	}

	// 3. 用户列出的 optional 模块（纯业务/集成，累加在 alwaysOn + optOut 之上）
	for _, raw := range c.Module {
		m := strings.ToLower(strings.TrimSpace(raw))
		if m == "" {
			continue
		}
		add(m)
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

// Package config 提供配置加载和管理功能
// 支持从YAML文件和环境变量加载配置，并支持模块化的配置管理
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构体，包含所有系统配置项
type Config struct {
	App      AppConfig      `yaml:"app"`      // 应用基础配置
	Database DatabaseConfig `yaml:"database"` // 数据库配置
	Redis    RedisConfig    `yaml:"redis"`    // Redis配置
	JWT      JWTConfig      `yaml:"jwt"`      // JWT认证配置
	Saas     SaasConfig     `yaml:"saas"`     // SaaS模式配置
	Log      LogConfig      `yaml:"log"`      // 日志配置
	Module   []string       `yaml:"module"`   // 启用的模块列表
	Apps     []string       `yaml:"apps"`     // 启用的应用列表
	Auth     AuthConfig     `yaml:"auth"`     // 认证配置
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name string `yaml:"name"` // 应用名称
	Env  string `yaml:"env"`  // 运行环境（dev/prod/test）
	Host string `yaml:"host"` // 服务器主机地址
	Port int    `yaml:"port"` // 服务器端口
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host               string `yaml:"host"`                   // 数据库主机地址
	Port               int    `yaml:"port"`                   // 数据库端口
	User               string `yaml:"user"`                   // 数据库用户名
	Password           string `yaml:"password"`               // 数据库密码
	DBName             string `yaml:"dbname"`                 // 数据库名称
	SSLMode            string `yaml:"sslmode"`                // SSL连接模式
	MaxOpenConns       int    `yaml:"max_open_conns"`         // 最大打开连接数
	MaxIdleConns       int    `yaml:"max_idle_conns"`         // 最大空闲连接数
	ConnMaxLifetimeSec int    `yaml:"conn_max_lifetime_sec"`  // 连接最大生命周期（秒）
	ConnMaxIdleTimeSec int    `yaml:"conn_max_idle_time_sec"` // 连接最大空闲时间（秒）
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host           string `yaml:"host"`             // Redis主机地址
	Port           int    `yaml:"port"`             // Redis端口
	Password       string `yaml:"password"`         // Redis密码
	DB             int    `yaml:"db"`               // Redis数据库编号
	Enabled        bool   `yaml:"enabled"`          // 是否启用Redis
	Required       bool   `yaml:"required"`         // Redis是否为必需（启动时检查）
	PoolSize       int    `yaml:"pool_size"`        // 连接池大小
	MinIdleConns   int    `yaml:"min_idle_conns"`   // 最小空闲连接数
	PoolTimeoutSec int    `yaml:"pool_timeout_sec"` // 连接池超时时间（秒）
	IdleTimeoutSec int    `yaml:"idle_timeout_sec"` // 空闲连接超时时间（秒）
	MaxConnAgeSec  int    `yaml:"max_conn_age_sec"` // 连接最大存活时间（秒）
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string `yaml:"secret"`         // JWT密钥
	Expire        int    `yaml:"expire"`         // Token过期时间（秒）
	RefreshExpire int    `yaml:"refresh_expire"` // Refresh Token过期时间（秒）
}

// SaasConfig SaaS模式配置
type SaasConfig struct {
	Mode string `yaml:"mode"` // SaaS模式（single/multi）
}

// LogConfig 日志配置
type LogConfig struct {
	Dir   string `yaml:"dir"`   // 日志目录
	Level string `yaml:"level"` // 日志级别（debug/info/warn/error）
}

// AuthConfig 认证配置
type AuthConfig struct {
	MaxLoginAttempts      int    `yaml:"max_login_attempts"`       // 最大登录尝试次数
	LockDurationSec       int    `yaml:"lock_duration_sec"`        // 账户锁定持续时间（秒）
	PasswordPolicy        string `yaml:"password_policy"`          // 密码策略
	TokenExpireSec        int    `yaml:"token_expire_sec"`         // Token过期时间（秒）
	RefreshTokenExpireSec int    `yaml:"refresh_token_expire_sec"` // Refresh Token过期时间（秒）
}

// DSN 生成PostgreSQL数据库连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// Addr 生成Redis连接地址（host:port格式）
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// cfg 全局配置实例
var cfg *Config

// Load 从指定路径加载配置文件
// 会先加载.env文件中的环境变量，然后加载YAML配置文件
// 最后用环境变量覆盖配置值，并验证模块配置的有效性
func Load(path string) (*Config, error) {
	if err := loadEnv(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env failed: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg = &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	overrideWithEnv(cfg)
	if err := validateModules(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadEnv 加载.env文件中的环境变量
// 只会设置当前未存在的环境变量，不会覆盖已有的环境变量
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

// overrideWithEnv 使用环境变量覆盖配置值
// 支持通过XIN_前缀的环境变量覆盖所有配置项
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

	envStr("XIN_SAAS_MODE", &c.Saas.Mode)

	envStr("XIN_LOG_DIR", &c.Log.Dir)
	envStr("XIN_LOG_LEVEL", &c.Log.Level)
	envCSV("XIN_MODULE", &c.Module)
}

// envStr 从环境变量读取字符串值并设置到目标变量
// 如果环境变量为空，则不修改目标变量
func envStr(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

// envInt 从环境变量读取整数值并设置到目标变量
// 如果环境变量为空或解析失败，则不修改目标变量
func envInt(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}

// envBool 从环境变量读取布尔值并设置到目标变量
// 如果环境变量为空或解析失败，则不修改目标变量
func envBool(key string, target *bool) {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			*target = b
		}
	}
}

// envCSV 从环境变量读取逗号分隔的字符串列表
// 会自动去除空格并转换为小写，忽略空值
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

// allowedModules 允许的模块白名单
var allowedModules = map[string]struct{}{
	"system": {}, // 系统模块
	"auth":   {}, // 认证模块
	"weixin": {}, // 微信模块
	"cms":    {}, // 内容管理模块
}

// validateModules 验证模块配置的有效性
// - 如果模块列表为空，默认启用system模块
// - 检查所有模块是否在白名单中
// - 去重并清理模块列表
func validateModules(c *Config) error {
	if len(c.Module) == 0 {
		c.Module = []string{"system"}
	}
	seen := map[string]struct{}{}
	for i := range c.Module {
		d := strings.ToLower(strings.TrimSpace(c.Module[i]))
		if d == "" {
			continue
		}
		if _, ok := allowedModules[d]; !ok {
			return fmt.Errorf("invalid module: %s (allowed: system,auth,weixin,cms)", d)
		}
		seen[d] = struct{}{}
	}
	if len(seen) == 0 {
		return errors.New("module is empty after validation")
	}
	c.Module = make([]string, 0, len(seen))
	for d := range seen {
		c.Module = append(c.Module, d)
	}
	return nil
}

// ModuleEnabled 检查指定模块是否已启用
// 模块名不区分大小写，会自动转换为小写进行比较
func (c *Config) ModuleEnabled(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, d := range c.Module {
		if d == name {
			return true
		}
	}
	return false
}

// AppEnabled 检查指定应用是否已启用
// 应用名不区分大小写，会自动转换为小写进行比较
func (c *Config) AppEnabled(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, a := range c.Apps {
		if strings.ToLower(strings.TrimSpace(a)) == name {
			return true
		}
	}
	return false
}

// Get 获取全局配置实例
// 如果配置尚未加载，返回nil
func Get() *Config {
	return cfg
}

// moduleBaseDir 模块配置文件的基准目录
var moduleBaseDir = filepath.Join("config", "modules")

// SetModuleBaseDir 设置模块配置文件的基准目录
// 用于自定义模块配置文件的存储位置
func SetModuleBaseDir(dir string) {
	moduleBaseDir = dir
}

// LoadModule 加载指定模块的配置文件
// 配置文件位于moduleBaseDir目录下，文件名为{name}.yaml
// 如果文件不存在或为空，不会报错，直接返回
// 加载后会使用环境变量覆盖配置值
func LoadModule(name string, target interface{}) error {
	path := filepath.Join(moduleBaseDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read module config %s: %w", name, err)
	}
	if len(data) == 0 {
		return nil
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse module config %s: %w", name, err)
	}
	overrideModuleEnv(name, target)
	return nil
}

// overrideModuleEnv 使用环境变量覆盖模块配置
// 环境变量命名规则：XIN_{MODULE_NAME}_{FIELD_NAME}
// 例如：XIN_AUTH_SECRET 会覆盖 auth 模块配置中的 secret 字段
// 支持字符串、整数、布尔值和浮点数类型
func overrideModuleEnv(module string, target interface{}) {
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

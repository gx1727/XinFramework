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

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Saas     SaasConfig     `yaml:"saas"`
	Log      LogConfig      `yaml:"log"`
	Module   []string       `yaml:"module"`
	Apps     []string       `yaml:"apps"`
	Auth     AuthConfig     `yaml:"auth"`
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

type SaasConfig struct {
	Mode string `yaml:"mode"`
}

type LogConfig struct {
	Dir   string `yaml:"dir"`
	Level string `yaml:"level"`
}

type AuthConfig struct {
	MaxLoginAttempts      int    `yaml:"max_login_attempts"`
	LockDurationSec       int    `yaml:"lock_duration_sec"`
	PasswordPolicy        string `yaml:"password_policy"`
	TokenExpireSec        int    `yaml:"token_expire_sec"`
	RefreshTokenExpireSec int    `yaml:"refresh_token_expire_sec"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

var cfg *Config

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

	envStr("XIN_SAAS_MODE", &c.Saas.Mode)

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

var allowedModules = map[string]struct{}{
	"system": {},
	"auth":   {},
	"weixin": {},
	"cms":    {},
}

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

func Get() *Config {
	return cfg
}

var moduleBaseDir = filepath.Join("config", "modules")

func SetModuleBaseDir(dir string) {
	moduleBaseDir = dir
}

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

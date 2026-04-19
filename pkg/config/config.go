package config

import (
	"bufio"
	"fmt"
	"os"
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
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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

	envStr("XIN_REDIS_HOST", &c.Redis.Host)
	envInt("XIN_REDIS_PORT", &c.Redis.Port)
	envStr("XIN_REDIS_PASSWORD", &c.Redis.Password)
	envInt("XIN_REDIS_DB", &c.Redis.DB)

	envStr("XIN_JWT_SECRET", &c.JWT.Secret)
	envInt("XIN_JWT_EXPIRE", &c.JWT.Expire)
	envInt("XIN_JWT_REFRESH_EXPIRE", &c.JWT.RefreshExpire)

	envStr("XIN_SAAS_MODE", &c.Saas.Mode)

	envStr("XIN_LOG_DIR", &c.Log.Dir)
	envStr("XIN_LOG_LEVEL", &c.Log.Level)
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

func Get() *Config {
	return cfg
}

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gx1727.com/xin/framework/pkg/config"
)

var Client *redis.Client

func Init(cfg *config.RedisConfig) error {
	if !cfg.Enabled {
		Client = nil
		return nil
	}

	opts := &redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}
	if cfg.PoolTimeoutSec > 0 {
		opts.PoolTimeout = time.Duration(cfg.PoolTimeoutSec) * time.Second
	}
	if cfg.IdleTimeoutSec > 0 {
		opts.IdleTimeout = time.Duration(cfg.IdleTimeoutSec) * time.Second
	}
	if cfg.MaxConnAgeSec > 0 {
		opts.MaxConnAge = time.Duration(cfg.MaxConnAgeSec) * time.Second
	}
	Client = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		if cfg.Required {
			return fmt.Errorf("redis ping failed: %w", err)
		}
		Client = nil
	}
	return nil
}

func Get() *redis.Client {
	return Client
}

func Close() error {
	if Client == nil {
		return nil
	}
	err := Client.Close()
	Client = nil
	return err
}

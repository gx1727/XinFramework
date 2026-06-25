// Package cache 提供 Redis 客户端的进程级单例与生命周期管理。
//
// 核心约束：
//   - Init / Get / Close 都依赖进程级 Client 变量（仅一份）
//   - 当 Redis 配置为未启用（enabled=false）或 Ping 失败且 required=false 时，
//     Client 为 nil，框架其他模块需自行 nil 检查并降级
//   - 当 Redis 配置为必启用（required=true）时 Ping 失败则 Init 返回 error
//
// 使用示例（cmd/xin/main.go 启动期）：
//
//	if err := cache.Init(cfg.Redis); err != nil {
//	    log.Fatalf("cache init failed: %v", err)
//	}
//	defer cache.Close()
//
//	// 业务代码取客户端
//	rdb := cache.Get()
//	if rdb == nil {
//	    // 降级到本地内存或 DB
//	}
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gx1727.com/xin/framework/pkg/config"
)

// Client 是进程级 Redis 客户端单例。可能为 nil（Redis 未启用或 Ping 失败）。
//
// 其他包应通过 Get() 获取，不要直接引用该变量。
var Client *redis.Client

// Init 根据配置初始化 Redis 客户端。
//
// 行为：
//   - enabled=false → Client=nil，nil error
//   - enabled=true + Ping 成功 → Client=已连上 Redis
//   - enabled=true + Ping 失败 + required=true → 返回 error
//   - enabled=true + Ping 失败 + required=false → Client=nil（静默降级）
//
// 失败降级由调用方通过 Get()==nil 自行处理。
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

// Get 返回当前 Redis 客户端单例。Redis 禁用或 Ping 失败时返回 nil。
//
// 业务代码必须先 nil 检查再使用，例如：
//
//	rdb := cache.Get()
//	if rdb == nil { return ErrRedisUnavailable }
//	return rdb.Set(ctx, key, val, ttl).Err()
func Get() *redis.Client {
	return Client
}

// Close 关闭 Redis 连接并清空 Client 变量。可重复调用。
func Close() error {
	if Client == nil {
		return nil
	}
	err := Client.Close()
	Client = nil
	return err
}

package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/permission"
)

const (
	permCacheKeyPrefix = "user:perm:"
	dsCacheKeyPrefix   = "user:ds:"
	permCacheTTL       = 15 * time.Minute
	dsCacheTTL         = 30 * time.Minute
)

// RedisPermissionCache implements permission.PermissionCache using Redis
type RedisPermissionCache struct{}

func NewRedisPermissionCache() *RedisPermissionCache {
	return &RedisPermissionCache{}
}

// GetPermissions retrieves cached permissions for a user
func (c *RedisPermissionCache) GetPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	rdb := cache.Get()
	if rdb == nil {
		return nil, nil // Cache unavailable
	}

	key := permCacheKeyPrefix + formatUint(userID)
	data, err := rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var perms map[string]bool
	if err := json.Unmarshal(data, &perms); err != nil {
		return nil, err
	}
	return perms, nil
}

// SetPermissions caches permissions for a user
func (c *RedisPermissionCache) SetPermissions(ctx context.Context, userID uint, perms map[string]bool) error {
	rdb := cache.Get()
	if rdb == nil {
		return nil
	}

	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}

	key := permCacheKeyPrefix + formatUint(userID)
	return rdb.Set(ctx, key, data, permCacheTTL).Err()
}

// InvalidatePermissions removes cached permissions for a user
func (c *RedisPermissionCache) InvalidatePermissions(ctx context.Context, userID uint) error {
	rdb := cache.Get()
	if rdb == nil {
		return nil
	}
	key := permCacheKeyPrefix + formatUint(userID)
	return rdb.Del(ctx, key).Err()
}

// GetDataScope retrieves cached data scope for a user
func (c *RedisPermissionCache) GetDataScope(ctx context.Context, userID uint) (*permission.DataScope, error) {
	rdb := cache.Get()
	if rdb == nil {
		return nil, nil
	}

	key := dsCacheKeyPrefix + formatUint(userID)
	data, err := rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var ds permission.DataScope
	if err := json.Unmarshal(data, &ds); err != nil {
		return nil, err
	}
	return &ds, nil
}

// SetDataScope caches data scope for a user
func (c *RedisPermissionCache) SetDataScope(ctx context.Context, userID uint, ds *permission.DataScope) error {
	rdb := cache.Get()
	if rdb == nil {
		return nil
	}

	data, err := json.Marshal(ds)
	if err != nil {
		return err
	}

	key := dsCacheKeyPrefix + formatUint(userID)
	return rdb.Set(ctx, key, data, dsCacheTTL).Err()
}

// InvalidateDataScope removes cached data scope for a user
func (c *RedisPermissionCache) InvalidateDataScope(ctx context.Context, userID uint) error {
	rdb := cache.Get()
	if rdb == nil {
		return nil
	}
	key := dsCacheKeyPrefix + formatUint(userID)
	return rdb.Del(ctx, key).Err()
}

func formatUint(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

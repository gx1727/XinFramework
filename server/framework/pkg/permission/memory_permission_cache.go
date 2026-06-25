package permission

import (
	"context"
	"sync"
	"time"
)

// MemoryPermissionCache 是进程内 in-memory 的 PermissionCache 实现。
//
// 用途：
//   - 当 Redis 不可用时（cache.Init 配置为 enabled=false 或 Ping 失败且 required=false），
//     boot.Init 会自动 fallback 到本实现——避免每次请求都查 DB
//   - 单元测试：不需要拉 Redis 容器就能验证 PermissionService 缓存逻辑
//
// 限制（与 Redis 实现的区别）：
//   - 数据不跨进程：每个进程有各自的缓存。多实例部署时仍会出现“某实例权限已
//     失效，其他实例还没感知”的情况，这是 in-memory 缓存的固有特性。生产
//     环境必须用 RedisPermissionCache。
//   - 内存增长无界：极端情况下（用户量极大且 InvalidateUser 频率低）可能占用
//     较多内存。可以通过降低 TTL 缓解。
type MemoryPermissionCache struct {
	mu sync.RWMutex

	permsTTL time.Duration
	dsTTL   time.Duration
	now     func() time.Time // 用于测试注入

	perms map[uint]memCacheEntry[map[string]bool]
	ds    map[uint]memCacheEntry[*DataScope]
}

type memCacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

// NewMemoryPermissionCache 构造 in-memory 权限缓存。
// 默认 TTL：权限 15 分钟、数据范围 30 分钟，与 Redis 实现保持一致。
func NewMemoryPermissionCache() *MemoryPermissionCache {
	return &MemoryPermissionCache{
		permsTTL: 15 * time.Minute,
		dsTTL:    30 * time.Minute,
		now:      time.Now,
		perms:    make(map[uint]memCacheEntry[map[string]bool]),
		ds:       make(map[uint]memCacheEntry[*DataScope]),
	}
}

// SetPermTTL 覆盖默认权限 TTL（用于测试或运维调优）。
func (c *MemoryPermissionCache) SetPermTTL(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.permsTTL = d
}

// SetDataScopeTTL 覆盖默认数据范围 TTL。
func (c *MemoryPermissionCache) SetDataScopeTTL(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dsTTL = d
}

// GetPermissions 取缓存的权限 map。未命中或已过期返回 (nil, nil)。
func (c *MemoryPermissionCache) GetPermissions(_ context.Context, userID uint) (map[string]bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.perms[userID]
	if !ok {
		return nil, nil
	}
	if c.now().After(entry.expiresAt) {
		return nil, nil
	}
	return entry.value, nil
}

// SetPermissions 写入权限缓存，覆盖已有条目。
func (c *MemoryPermissionCache) SetPermissions(_ context.Context, userID uint, perms map[string]bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.perms[userID] = memCacheEntry[map[string]bool]{
		value:     perms,
		expiresAt: c.now().Add(c.permsTTL),
	}
	return nil
}

// InvalidatePermissions 删除指定用户的权限缓存条目。
func (c *MemoryPermissionCache) InvalidatePermissions(_ context.Context, userID uint) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.perms, userID)
	return nil
}

// GetDataScope 取缓存的数据范围。未命中或已过期返回 (nil, nil)。
func (c *MemoryPermissionCache) GetDataScope(_ context.Context, userID uint) (*DataScope, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.ds[userID]
	if !ok {
		return nil, nil
	}
	if c.now().After(entry.expiresAt) {
		return nil, nil
	}
	return entry.value, nil
}

// SetDataScope 写入数据范围缓存。
func (c *MemoryPermissionCache) SetDataScope(_ context.Context, userID uint, ds *DataScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ds[userID] = memCacheEntry[*DataScope]{
		value:     ds,
		expiresAt: c.now().Add(c.dsTTL),
	}
	return nil
}

// InvalidateDataScope 删除指定用户的数据范围缓存。
func (c *MemoryPermissionCache) InvalidateDataScope(_ context.Context, userID uint) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ds, userID)
	return nil
}

// Compile-time guarantee: MemoryPermissionCache satisfies PermissionCache.
var _ PermissionCache = (*MemoryPermissionCache)(nil)
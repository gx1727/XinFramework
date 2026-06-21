// Package config 通用配置 - 进程内合并缓存
//
// 与 dict 的 framework/pkg/dict 缓存对标：
//   - 按 (tenantID, groupCode) 缓存 ResolvedConfig（含 platform + override 合并结果）
//   - 任何 group 修改触发 Invalidate(tenantID) → 整体失效
//   - 懒加载：第一次访问时从 DB 加载并填充
//
// 为什么按整体 tenant 失效而非单 group 失效：
//   - 平台 group 的修改对所有租户都可见（visibility=all 时）
//   - 单 group 失效不能覆盖跨租户场景
//   - 配置数据量小（一般 < 1000 项/租户），整体失效代价低
//
// 平台 group CRUD 不经过这个缓存（走 PlatformService 单独路径）。
package config

import (
	"sync"
)

// Cache 进程内合并缓存
//
// key = uint(tenantID)
// value = map[groupCode]*ResolvedConfig（业务合并后的最终视图）
type Cache struct {
	m sync.Map
}

func NewCache() *Cache { return &Cache{} }

// Get 取合并后的 ResolvedConfig
func (c *Cache) Get(tenantID uint, groupCode string) (*ResolvedConfig, bool) {
	v, ok := c.m.Load(tenantID)
	if !ok {
		return nil, false
	}
	groups, ok := v.(map[string]*ResolvedConfig)
	if !ok {
		return nil, false
	}
	rc, ok := groups[groupCode]
	return rc, ok
}

// GetAll 取某租户的全部合并配置
func (c *Cache) GetAll(tenantID uint) (map[string]*ResolvedConfig, bool) {
	v, ok := c.m.Load(tenantID)
	if !ok {
		return nil, false
	}
	groups, ok := v.(map[string]*ResolvedConfig)
	return groups, ok
}

// Put 写入合并后的配置
//
// 应在 Service.ResolveAllForTenant 后调用，一次性写满。
func (c *Cache) Put(tenantID uint, groups map[string]*ResolvedConfig) {
	c.m.Store(tenantID, groups)
}

// Invalidate 失效某租户的全部缓存
//
// 触发时机：
//   - 租户改了 config_items.value
//   - 租户 upsert/delete 了 override
//   - 平台改了 platform item（影响所有租户 — 遍历失效）
//   - super_admin 改了 visibility
func (c *Cache) Invalidate(tenantID uint) {
	c.m.Delete(tenantID)
}

// InvalidateAll 失效所有租户的缓存（平台级改动）
//
// 遍历所有 key 逐个删除。配置数据量小，可以 O(N)。
func (c *Cache) InvalidateAll() {
	c.m.Range(func(key, _ any) bool {
		c.m.Delete(key)
		return true
	})
}

// LoadOrLoadAll 懒加载模式：若租户无缓存，加载 fn 后 Put，返回结果
//
// fn 应该是 Service.ResolveAllForTenant 的实现。
// 用 callback 避免 cache 直接依赖 service（防止循环引用）。
func (c *Cache) LoadOrLoadAll(tenantID uint, fn func() (map[string]*ResolvedConfig, error)) (map[string]*ResolvedConfig, error) {
	if groups, ok := c.GetAll(tenantID); ok {
		return groups, nil
	}
	groups, err := fn()
	if err != nil {
		return nil, err
	}
	c.Put(tenantID, groups)
	return groups, nil
}

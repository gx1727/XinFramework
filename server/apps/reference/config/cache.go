// Package config 通用配置 - 进程内缓存
package config

import "sync"

// Cache 进程内缓存（按 tenant 分组缓存 items 列表）
type Cache struct {
	m sync.Map // key=uint(tenantID), value=map[string][]*ConfigItem (groupCode -> items)
}

func NewCache() *Cache { return &Cache{} }

func (c *Cache) GetAll(tenantID uint) (map[string][]*ConfigItem, bool) {
	v, ok := c.m.Load(tenantID)
	if !ok {
		return nil, false
	}
	groups, ok := v.(map[string][]*ConfigItem)
	return groups, ok
}

func (c *Cache) PutAll(tenantID uint, groups map[string][]*ConfigItem) {
	c.m.Store(tenantID, groups)
}

func (c *Cache) Invalidate(tenantID uint) {
	c.m.Delete(tenantID)
}

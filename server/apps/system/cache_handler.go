package system

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/resp"
)

// CacheInfo 获取 Redis 服务器状态信息
func (h *Handler) CacheInfo(c *gin.Context) {
	client := cache.Get()
	if client == nil {
		resp.BadRequest(c, "Redis cache is not enabled")
		return
	}

	ctx := context.Background()

	// 获取 INFO
	infoStr, err := client.Info(ctx).Result()
	if err != nil {
		resp.ServerError(c, "Failed to get redis info: "+err.Error())
		return
	}

	// 解析 info 字符串
	info := make(map[string]string)
	lines := strings.Split(infoStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			info[parts[0]] = parts[1]
		}
	}

	// 获取 DBSize
	dbSize, _ := client.DBSize(ctx).Result()

	// 获取 CommandStats
	cmdStatsStr, _ := client.Info(ctx, "commandstats").Result()
	commandStats := make(map[string]string)
	cmdLines := strings.Split(cmdStatsStr, "\n")
	for _, line := range cmdLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			cmdName := strings.TrimPrefix(parts[0], "cmdstat_")
			commandStats[cmdName] = parts[1]
		}
	}

	result := gin.H{
		"info":         info,
		"dbSize":       dbSize,
		"commandStats": commandStats,
	}

	resp.Success(c, result)
}

// GetCacheKeys 根据模式分页获取键名列表。
//   - pattern：Redis MATCH 表达式，默认 "*"
//   - page：从 1 开始，默认 1
//   - size：每页条数，默认 50，上限 200
//   - exclude_prefixes：逗号分隔的前缀列表（如 "cache_,sess:"），
//     SCAN 全量后过滤掉以任一前缀开头的 key。空值 = 不过滤。
//
// 实现：用 SCAN 迭代代替 KEYS（不阻塞 Redis），全量收集后 sort 保持分页稳定，
// 再按 (page, size) 切片返回 {list, total}。过滤在 sort 之前进行。
func (h *Handler) GetCacheKeys(c *gin.Context) {
	pattern := c.Query("pattern")
	if pattern == "" {
		pattern = "*"
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	if size < 1 || size > 200 {
		size = 50
	}

	var excludePrefixes []string
	if raw := c.Query("exclude_prefixes"); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				excludePrefixes = append(excludePrefixes, p)
			}
		}
	}

	client := cache.Get()
	if client == nil {
		resp.BadRequest(c, "Redis cache is not enabled")
		return
	}

	ctx := context.Background()
	var (
		allKeys []string
		cursor  uint64
	)
	for {
		keys, next, err := client.Scan(ctx, cursor, pattern, 500).Result()
		if err != nil {
			resp.ServerError(c, "Failed to scan keys: "+err.Error())
			return
		}
		allKeys = append(allKeys, keys...)
		if next == 0 {
			break
		}
		cursor = next
	}
	sort.Strings(allKeys)

	if len(excludePrefixes) > 0 {
		filtered := make([]string, 0, len(allKeys))
		for _, k := range allKeys {
			skip := false
			for _, p := range excludePrefixes {
				if strings.HasPrefix(k, p) {
					skip = true
					break
				}
			}
			if !skip {
				filtered = append(filtered, k)
			}
		}
		allKeys = filtered
	}

	total := int64(len(allKeys))
	start := (page - 1) * size
	if start > int(total) {
		start = int(total)
	}
	end := start + size
	if end > int(total) {
		end = int(total)
	}
	list := allKeys[start:end]

	resp.Paginate(c, total, list)
}

// GetCacheValue 获取特定键的值和 TTL
func (h *Handler) GetCacheValue(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" {
		resp.BadRequest(c, "Key is required")
		return
	}

	client := cache.Get()
	if client == nil {
		resp.BadRequest(c, "Redis cache is not enabled")
		return
	}

	ctx := context.Background()

	// 获取类型
	keyType, err := client.Type(ctx, key).Result()
	if err != nil {
		resp.ServerError(c, "Failed to get key type: "+err.Error())
		return
	}

	if keyType == "none" {
		resp.BadRequest(c, "Key does not exist")
		return
	}

	var value any
	switch keyType {
	case "string":
		value, _ = client.Get(ctx, key).Result()
	case "hash":
		value, _ = client.HGetAll(ctx, key).Result()
	case "list":
		value, _ = client.LRange(ctx, key, 0, -1).Result()
	case "set":
		value, _ = client.SMembers(ctx, key).Result()
	case "zset":
		value, _ = client.ZRange(ctx, key, 0, -1).Result()
	default:
		value = "unsupported type: " + keyType
	}

	ttl, _ := client.TTL(ctx, key).Result()

	resp.Success(c, gin.H{
		"key":   key,
		"type":  keyType,
		"value": value,
		"ttl":   ttl.Seconds(),
	})
}

// DeleteCacheKey 删除特定键
func (h *Handler) DeleteCacheKey(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" {
		resp.BadRequest(c, "Key is required")
		return
	}

	client := cache.Get()
	if client == nil {
		resp.BadRequest(c, "Redis cache is not enabled")
		return
	}

	ctx := context.Background()
	err := client.Del(ctx, key).Err()
	if err != nil {
		resp.ServerError(c, "Failed to delete key: "+err.Error())
		return
	}

	resp.Success(c, "Key deleted successfully")
}

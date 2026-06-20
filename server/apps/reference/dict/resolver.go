// Package dict 跨租户字典解析器（业务最终消费入口）
//
// 设计目标：
//   - 业务代码只调这一个包，就能拿到"租户视角下合并后的字典"
//   - 内部走 Service → Repository 的合并逻辑（COALESCE 覆盖）
//   - 提供单条/批量两种调用形态，避免 N+1
package dict

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Resolver 对外暴露的解析器（其他模块通过 plugin.AppContext.DictResolver() 拿到）
type Resolver struct {
	pool *pgxpool.Pool
	repo DictRepository
}

// NewResolver 构造 Resolver
func NewResolver(pool *pgxpool.Pool, repo DictRepository) *Resolver {
	if repo == nil {
		repo = NewPostgresDictRepository(pool)
	}
	return &Resolver{pool: pool, repo: repo}
}

// Resolve 单条：按 dict code 取租户视角的合并字典
func (r *Resolver) Resolve(ctx context.Context, tenantID uint, dictCode string) (*ResolvedDict, error) {
	if tenantID == 0 {
		return nil, ErrDictInvisible
	}
	if dictCode == "" {
		return nil, fmt.Errorf("dict code required")
	}
	return r.repo.ResolveDictForTenant(ctx, tenantID, dictCode)
}

// ResolveBatch 批量：按一组 code 取租户视角的合并字典（避免 N+1）
// 返回 map[code]ResolvedDict，code 不存在 / 不可见 不会出现在 map 里
func (r *Resolver) ResolveBatch(ctx context.Context, tenantID uint, codes []string) (map[string]*ResolvedDict, error) {
	if tenantID == 0 {
		return nil, ErrDictInvisible
	}
	if len(codes) == 0 {
		return map[string]*ResolvedDict{}, nil
	}

	out := make(map[string]*ResolvedDict, len(codes))
	for _, code := range codes {
		rd, err := r.repo.ResolveDictForTenant(ctx, tenantID, code)
		if err != nil {
			// 不可见 / 不存在 → 跳过；其它错误才返回
			continue
		}
		out[code] = rd
	}
	return out, nil
}

// ResolveParallel 批量并发解析（高频场景用）
func (r *Resolver) ResolveParallel(ctx context.Context, tenantID uint, codes []string) (map[string]*ResolvedDict, error) {
	if tenantID == 0 {
		return nil, ErrDictInvisible
	}
	if len(codes) == 0 {
		return map[string]*ResolvedDict{}, nil
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		out     = make(map[string]*ResolvedDict, len(codes))
		firstErr error
	)

	// 限制并发度（防止大 code 列表打爆 DB）
	sem := make(chan struct{}, 8)

	for _, code := range codes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			defer func() { <-sem }()

			rd, err := r.repo.ResolveDictForTenant(ctx, tenantID, c)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			out[c] = rd
			mu.Unlock()
		}(code)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}
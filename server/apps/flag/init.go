package flag

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"gx1727.com/xin/framework/pkg/config"
)

// 包级变量 - 全局 Repository 实例
var (
	frameRepo     *FrameRepository
	avatarRepo    *AvatarRepository
	frameCatRepo  *FrameCategoryRepository
	avatarCatRepo *AvatarCategoryRepository
)

// 包级变量 - DB pool / config 引用，由 module.go 在 Module(app) 里注入。
// 之所以用包级变量：flag 的 handler 数量很多（>20 处 RunInTenantTx），
// 让它们全部接收 pool 参数会改几十处，所以采用 module.go 注入一次、
// handler 共享的过渡写法。未来应改为 handler 持有 pool 字段。
var (
	dbPool    *pgxpool.Pool
	cfgRef    *config.Config
)

// InitRepositories 初始化所有 Repository（在模块启动时调用）
func InitRepositories(pool *pgxpool.Pool) {
	dbPool = pool
	frameRepo = NewFrameRepository(pool)
	avatarRepo = NewAvatarRepository(pool)
	frameCatRepo = NewFrameCategoryRepository(pool)
	avatarCatRepo = NewAvatarCategoryRepository(pool)
}

// SetConfig 注入全局 config 引用，供 handler 在 GenerateAvatar 等场景读 storage 配置。
func SetConfig(cfg *config.Config) {
	cfgRef = cfg
}

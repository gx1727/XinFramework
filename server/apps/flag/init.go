package flag

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// 包级变量 - 全局 Repository 实例
var (
	frameRepo     *FrameRepository
	avatarRepo    *AvatarRepository
	frameCatRepo  *FrameCategoryRepository
	avatarCatRepo *AvatarCategoryRepository
)

// InitRepositories 初始化所有 Repository（在模块启动时调用）
func InitRepositories(pool *pgxpool.Pool) {
	frameRepo = NewFrameRepository(pool)
	avatarRepo = NewAvatarRepository(pool)
	frameCatRepo = NewFrameCategoryRepository(pool)
	avatarCatRepo = NewAvatarCategoryRepository(pool)
}

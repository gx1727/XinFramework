package cms

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
)

// CmsConfig CMS模块配置结构
type CmsConfig struct {
	PostPerPage   int    `yaml:"post_per_page"`   // 每页文章数量
	UploadMaxSize int64  `yaml:"upload_max_size"` // 上传文件最大大小（字节）
	UploadDir     string `yaml:"upload_dir"`      // 上传文件目录
}

// moduleCfg 全局CMS配置实例
var moduleCfg *CmsConfig

// Cfg 获取CMS模块配置
func Cfg() *CmsConfig {
	return moduleCfg
}

// Register 注册CMS模块的API路由
func Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	// 注册受保护的ping接口，用于健康检查和配置查看
	protected.GET("/cms/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{
			"domain": "cms",
			"status": "enabled",
			"config": moduleCfg,
		})
	})
}

// Module 创建并返回CMS插件模块实例
func Module() plugin.Module {
	return plugin.NewModuleWithOpts("cms", Register,
		plugin.WithInit(initModule),       // 设置初始化函数
		plugin.WithMigrate(migrateModule), // 设置迁移函数
	)
}

// initModule 初始化CMS模块配置
func initModule() error {
	// 设置默认配置
	moduleCfg = &CmsConfig{
		PostPerPage:   20,
		UploadMaxSize: 10 << 20, // 10MB
		UploadDir:     "uploads/cms",
	}

	// 尝试从配置文件加载配置
	cfgPath := configPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	// 解析YAML配置文件
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, moduleCfg); err != nil {
			return err
		}
	}
	return nil
}

// configPath 获取CMS配置文件路径
// 优先级：环境变量 XIN_CMS_CONFIG > 开发环境路径 > 生产环境路径
func configPath() string {
	// 检查环境变量
	if p := os.Getenv("XIN_CMS_CONFIG"); p != "" {
		return p
	}
	// 开发环境路径
	dev := filepath.Join("apps", "cms", "config.yaml")
	if _, err := os.Stat(dev); err == nil {
		return dev
	}
	// 生产环境路径
	return filepath.Join("config", "cms", "config.yaml")
}

// migrateModule 执行CMS模块的数据库迁移
func migrateModule() error {
	// 优先使用开发环境的迁移文件路径
	dev := filepath.Join("apps", "cms", "migrations")
	if _, err := os.Stat(dev); err == nil {
		return migrate.Run(dev)
	}
	// 使用发布版的迁移文件路径
	return migrate.Run(filepath.Join("migrations", "cms"))
}

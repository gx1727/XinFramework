package cms

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/migrate"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/resp"
)

type CmsConfig struct {
	PostPerPage   int    `yaml:"post_per_page"`
	UploadMaxSize int64  `yaml:"upload_max_size"`
	UploadDir     string `yaml:"upload_dir"`
}

var moduleCfg *CmsConfig

func Cfg() *CmsConfig {
	return moduleCfg
}

func Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	protected.GET("/cms/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{
			"domain": "cms",
			"status": "enabled",
			"config": moduleCfg,
		})
	})
}

func Module() plugin.Module {
	return plugin.NewModuleWithOpts("cms", Register,
		plugin.WithInit(initModule),
		plugin.WithMigrate(migrateModule),
	)
}

func initModule() error {
	moduleCfg = &CmsConfig{
		PostPerPage:   20,
		UploadMaxSize: 10 << 20,
		UploadDir:     "uploads/cms",
	}
	return config.LoadModule("cms", moduleCfg)
}

func migrateModule() error {
	dev := filepath.Join("apps", "cms", "migrations")
	if _, err := filepath.Abs(dev); err == nil {
		if _, err := filepath.Glob(dev); err == nil {
			return migrate.Run(dev)
		}
	}
	return migrate.Run(filepath.Join("migrations", "cms"))
}

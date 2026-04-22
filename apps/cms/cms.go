package cms

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
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

func RegisterV1(public, protected *gin.RouterGroup) {
	protected.GET("/cms/ping", func(c *gin.Context) {
		resp.Success(c, gin.H{
			"domain": "cms",
			"status": "enabled",
			"config": moduleCfg,
		})
	})
}

func Module() plugin.Module {
	return plugin.NewModuleWithOpts("cms", RegisterV1,
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

	cfgPath := configPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, moduleCfg); err != nil {
			return err
		}
	}
	return nil
}

func migrateModule() error {
	return migrate.Run(filepath.Join("apps", "cms", "migrations"))
}

func configPath() string {
	if p := os.Getenv("XIN_CMS_CONFIG"); p != "" {
		return p
	}
	return filepath.Join("apps", "cms", "config.yaml")
}

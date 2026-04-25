package cms

import (
	"gx1727.com/xin/framework/pkg/config"
)

type Config struct {
	PostPerPage   int    `yaml:"post_per_page"`
	UploadMaxSize int64  `yaml:"upload_max_size"`
	UploadDir     string `yaml:"upload_dir"`
}

var cfg *Config

func Cfg() *Config { return cfg }

func LoadConfig() error {
	cfg = &Config{
		PostPerPage:   20,
		UploadMaxSize: 10 << 20,
		UploadDir:     "uploads/cms",
	}
	return config.LoadModule("cms", cfg)
}

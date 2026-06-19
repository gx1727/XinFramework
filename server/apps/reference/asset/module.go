package asset

import (
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

// init 在包加载时自动注册到 plugin.Apps()。cmd/xin 通过 side-effect
// import 引入此包，避免 framework.go 中维护硬编码列表。
func init() {
	plugin.Register(Module())
}

// Module 返回 asset 模块的完整定义
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "asset",
		RegFn: func(_ plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			// Phase 4: config.Get()/db.Get() → bootx.Config()/bootx.Pool()
			cfg := bootx.Config()
			pool := bootx.Pool()
			// 创建 storage
			var s storage.Storage
			if cfg.Storage.Provider == "cos" {
				cosStorage, err := storage_cos.NewCosStorage(storage_cos.Config{
					URL:       cfg.Storage.CosURL,
					SecretID:  cfg.Storage.CosSecretID,
					SecretKey: cfg.Storage.CosSecretKey,
					BaseURL:   cfg.Storage.CosBaseURL,
				})
				if err != nil {
					log.Fatalf("failed to init cos storage: %v", err)
				}
				s = cosStorage
			} else {
				s = storage_local.NewLocalStorage(
					cfg.Storage.LocalDir,
					cfg.Storage.LocalBaseURL,
				)
			}

			// 创建 service 和 handler
			svc := NewFileService(s, NewAttachmentRepository(pool))
			h := NewFileHandler(svc)

			// 注册路由
			Register(public, protected, h)
		},
	}
}

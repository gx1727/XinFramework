package asset

import (
	"log"

	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

// Module 返回 asset 模块的完整定义
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "asset",
		RegFn: func(ctx plugin.Reader, slots plugin.RouterSlots) {
			public := slots.MustGet(plugin.SlotPublic).Group
			protected := slots.MustGet(plugin.SlotProtected).Group
			cfg := app.Config
			pool := app.DB.Raw()
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
